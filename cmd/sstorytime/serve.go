package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/markburgess/SSTorytime/internal/db"
	"github.com/markburgess/SSTorytime/internal/server"
	"github.com/spf13/cobra"
)

var (
	serveHTTPAddr  string
	serveHTTPSAddr string
	serveResources string
	serveCertFile  string
	serveKeyFile   string
	serveHTTPOnly  bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "HTTP UI and JSON API (TLS with auto self-signed certs)",
	Long: `Start the SSTorytime web UI and JSON API.

By default this matches upstream behaviour:
  - HTTPS on :8443 (API + static UI)
  - HTTP on :8080 redirects to HTTPS

If cert.pem / key.pem are missing, a self-signed localhost certificate is
generated with the Go standard library (crypto/x509) — no openssl required.
Browsers will warn on self-signed certs; accept the exception for local use.

Use --http-only for plain HTTP (e.g. behind a reverse proxy that terminates TLS).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if _, _, err := db.OpenDSN(ctx, databaseURL); err != nil {
			return fmt.Errorf("database: %w", err)
		}

		publicFS, err := server.PublicFS()
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		server.VERBOSE = verbose
		server.RegisterRoutes(mux, publicFS, serveResources)

		baseCtx := func(net.Listener) context.Context { return ctx }

		if serveHTTPOnly {
			srv := &http.Server{Addr: serveHTTPAddr, Handler: mux, BaseContext: baseCtx}
			go func() {
				log.Printf("listening on http://%s (HTTP only)", serveHTTPAddr)
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatal(err)
				}
			}()
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()
			return srv.Shutdown(shutdownCtx)
		}

		certFile, keyFile, generated, err := server.EnsureTLSMaterial(serveCertFile, serveKeyFile)
		if err != nil {
			return fmt.Errorf("tls: %w", err)
		}
		if generated {
			log.Printf("generated self-signed TLS material: %s , %s (dev only; browser will warn)", certFile, keyFile)
		}

		httpsSrv := &http.Server{Addr: serveHTTPSAddr, Handler: mux, BaseContext: baseCtx}
		httpSrv := &http.Server{
			Addr: serveHTTPAddr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				host := r.Host
				if h, _, err := net.SplitHostPort(host); err == nil {
					host = h
				} else {
					host = strings.TrimSuffix(host, ".")
				}
				// Preserve non-default HTTPS port in redirect.
				targetHost := host
				if _, port, err := net.SplitHostPort(serveHTTPSAddr); err == nil && port != "443" {
					targetHost = net.JoinHostPort(host, port)
				}
				dest := "https://" + targetHost + r.URL.RequestURI()
				http.Redirect(w, r, dest, http.StatusMovedPermanently)
			}),
			BaseContext: baseCtx,
		}

		go func() {
			log.Printf("listening on http://%s (redirect → HTTPS)", serveHTTPAddr)
			if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}()
		go func() {
			log.Printf("listening on https://%s", serveHTTPSAddr)
			if err := httpsSrv.ListenAndServeTLS(certFile, keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}()

		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		errHTTP := httpSrv.Shutdown(shutdownCtx)
		errHTTPS := httpsSrv.Shutdown(shutdownCtx)
		if errHTTPS != nil {
			return errHTTPS
		}
		return errHTTP
	},
}

func init() {
	registerMultiCall("serve", "http_server")
	serveCmd.Flags().StringVar(&serveHTTPAddr, "http-addr", ":8080", "HTTP listen address (redirects to HTTPS unless --http-only)")
	serveCmd.Flags().StringVar(&serveHTTPSAddr, "https-addr", ":8443", "HTTPS listen address")
	// Keep --addr as alias for https-addr for earlier PR flag name.
	serveCmd.Flags().StringVar(&serveHTTPSAddr, "addr", ":8443", "alias for --https-addr")
	serveCmd.Flags().StringVar(&serveResources, "resources", "/mnt", "directory for /Resources/")
	serveCmd.Flags().StringVar(&serveCertFile, "cert", "cert.pem", "TLS certificate PEM (created if missing)")
	serveCmd.Flags().StringVar(&serveKeyFile, "key", "key.pem", "TLS private key PEM (created if missing)")
	serveCmd.Flags().BoolVar(&serveHTTPOnly, "http-only", false, "serve plain HTTP only (no TLS, no redirect)")
}
