package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/markburgess/SSTorytime/internal/db"
	"github.com/markburgess/SSTorytime/internal/server"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	serveAddr      string
	serveHTTPAddr  string
	serveHTTPSAddr string
	serveResources string
	serveCertFile  string
	serveKeyFile   string
	serveTLS       bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "HTTP UI and JSON API",
	Long: `Start the SSTorytime web UI and JSON API.

Default is plain HTTP on --addr (:8080). That is what you want behind a
reverse proxy that terminates TLS (and handles ACME). The app does not speak ACME.

Local HTTPS (upstream-like) is opt-in:

  sstorytime serve --tls

Then HTTPS listens on --https-addr (:8443) and HTTP on --http-addr (:8080)
redirects to it. Missing cert.pem/key.pem are generated with crypto/x509
(self-signed localhost; browsers will warn).`,
	RunE: runServe,
}

func init() {
	registerMultiCall("serve", "http_server")
	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "HTTP listen address (app traffic; default plain HTTP)")
	serveCmd.Flags().BoolVar(&serveTLS, "tls", false, "enable self-signed HTTPS + HTTP→HTTPS redirect (local dev)")
	serveCmd.Flags().StringVar(&serveHTTPSAddr, "https-addr", ":8443", "HTTPS listen address (with --tls)")
	serveCmd.Flags().StringVar(&serveHTTPAddr, "http-addr", "", "HTTP redirect listen address with --tls (default: same as --addr)")
	serveCmd.Flags().StringVar(&serveResources, "resources", "/mnt", "directory for /Resources/")
	serveCmd.Flags().StringVar(&serveCertFile, "cert", "cert.pem", "TLS certificate PEM (created if missing; --tls only)")
	serveCmd.Flags().StringVar(&serveKeyFile, "key", "key.pem", "TLS private key PEM (created if missing; --tls only)")
}

func runServe(cmd *cobra.Command, args []string) error {
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

	if !serveTLS {
		return ServePlainHTTP(ctx, serveAddr, mux, baseCtx)
	}

	redirectAddr := serveHTTPAddr
	if redirectAddr == "" {
		redirectAddr = serveAddr
	}
	return ServeLocalTLS(ctx, redirectAddr, serveHTTPSAddr, mux, baseCtx)
}

// ServePlainHTTP listens for app traffic over HTTP (reverse-proxy default).
func ServePlainHTTP(ctx context.Context, addr string, mux http.Handler, baseCtx func(net.Listener) context.Context) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	srv := &http.Server{Handler: mux, BaseContext: baseCtx}
	log.Printf("listening on http://%s", ln.Addr())

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := srv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})
	return g.Wait()
}

// ServeLocalTLS enables self-signed HTTPS plus HTTP→HTTPS redirect for local dev.
func ServeLocalTLS(ctx context.Context, httpAddr, httpsAddr string, mux http.Handler, baseCtx func(net.Listener) context.Context) error {
	certFile, keyFile, generated, err := server.EnsureTLSMaterial(serveCertFile, serveKeyFile)
	if err != nil {
		return fmt.Errorf("tls: %w", err)
	}
	if generated {
		log.Printf("generated self-signed TLS material: %s , %s (dev only; browser will warn)", certFile, keyFile)
	}

	httpsLn, err := net.Listen("tcp", httpsAddr)
	if err != nil {
		return fmt.Errorf("listen https %s: %w", httpsAddr, err)
	}
	httpLn, err := net.Listen("tcp", httpAddr)
	if err != nil {
		if cerr := httpsLn.Close(); cerr != nil {
			return fmt.Errorf("listen http %s: %w (also close https: %v)", httpAddr, err, cerr)
		}
		return fmt.Errorf("listen http %s: %w", httpAddr, err)
	}

	httpsSrv := &http.Server{Handler: mux, BaseContext: baseCtx}
	httpSrv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}
			targetHost := host
			// Redirect to the actual HTTPS listener address port when not 443.
			if ta, ok := httpsLn.Addr().(*net.TCPAddr); ok && ta.Port != 443 {
				targetHost = net.JoinHostPort(host, fmt.Sprintf("%d", ta.Port))
			} else if _, port, err := net.SplitHostPort(httpsAddr); err == nil && port != "443" && port != "" {
				targetHost = net.JoinHostPort(host, port)
			}
			http.Redirect(w, r, "https://"+targetHost+r.URL.RequestURI(), http.StatusMovedPermanently)
		}),
		BaseContext: baseCtx,
	}

	log.Printf("listening on https://%s", httpsLn.Addr())
	log.Printf("listening on http://%s (redirect → HTTPS)", httpLn.Addr())

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := httpsSrv.ServeTLS(httpsLn, certFile, keyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("https: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		err := httpSrv.Serve(httpLn)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http redirect: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		var wg sync.WaitGroup
		wg.Add(2)
		var e1, e2 error
		go func() { defer wg.Done(); e1 = httpSrv.Shutdown(shutdownCtx) }()
		go func() { defer wg.Done(); e2 = httpsSrv.Shutdown(shutdownCtx) }()
		wg.Wait()
		return errors.Join(e1, e2)
	})
	return g.Wait()
}
