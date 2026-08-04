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

Default is a single plain-HTTP listener on --addr (:8080). That is what you
want behind a reverse proxy (proxy terminates TLS / ACME). The app does not
speak ACME and does not open a second port.

Local HTTPS is opt-in and still one port by default:

  sstorytime serve --tls
  # HTTPS only on --https-addr (:8443); self-signed cert if missing

Optional HTTP→HTTPS redirect (second port — only if you need it and the
HTTPS port is reachable):

  sstorytime serve --tls --http-addr :8080
`,
	RunE: runServe,
}

func init() {
	registerMultiCall("serve", "http_server")
	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "HTTP listen address when not using --tls")
	serveCmd.Flags().BoolVar(&serveTLS, "tls", false, "serve HTTPS (self-signed if cert/key missing); single port unless --http-addr is set")
	serveCmd.Flags().StringVar(&serveHTTPSAddr, "https-addr", ":8443", "HTTPS listen address (with --tls)")
	serveCmd.Flags().StringVar(&serveHTTPAddr, "http-addr", "", "optional second listener: HTTP→HTTPS redirect (with --tls only; omit to avoid dual bind)")
	serveCmd.Flags().StringVar(&serveResources, "resources", "/mnt", "directory for /Resources/")
	serveCmd.Flags().StringVar(&serveCertFile, "cert", "cert.pem", "TLS certificate PEM (created if missing; --tls only)")
	serveCmd.Flags().StringVar(&serveKeyFile, "key", "key.pem", "TLS private key PEM (created if missing; --tls only)")
}

func runServe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if _, _, err := db.OpenDSN(ctx, databaseURL); err != nil {
		return fmt.Errorf("database: %w", err)
	}
	// One graph session for all HTTP handlers (avoids per-request Open + cache reload hang).
	server.WarmSession(ctx)

	publicFS, err := server.PublicFS()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	server.VERBOSE = verbose
	server.RegisterRoutes(mux, publicFS, serveResources)

	// Do not tie request contexts to the process signal ctx as BaseContext parent
	// for cancellation semantics of in-flight work; Shutdown still stops the server.
	baseCtx := func(net.Listener) context.Context { return context.Background() }

	if !serveTLS {
		return ServePlainHTTP(ctx, serveAddr, mux, baseCtx)
	}

	// Redirect listener only when explicitly requested. Auto dual-bind was hanging
	// clients behind firewalls that allow :8080 but drop :8443 after a 301.
	redirectAddr := ""
	if cmd.Flags().Changed("http-addr") && serveHTTPAddr != "" {
		redirectAddr = serveHTTPAddr
	}
	return ServeLocalTLS(ctx, redirectAddr, serveHTTPSAddr, mux, baseCtx)
}

// ServePlainHTTP listens for app traffic over HTTP (reverse-proxy default).
func ServePlainHTTP(ctx context.Context, addr string, mux http.Handler, baseCtx func(net.Listener) context.Context) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	srv := &http.Server{
		Handler:           mux,
		BaseContext:       baseCtx,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
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

// ServeLocalTLS serves the app over HTTPS. If redirectHTTPAddr is non-empty,
// a second HTTP listener only issues redirects to HTTPS (opt-in dual bind).
func ServeLocalTLS(ctx context.Context, redirectHTTPAddr, httpsAddr string, mux http.Handler, baseCtx func(net.Listener) context.Context) error {
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

	httpsSrv := &http.Server{
		Handler:           mux,
		BaseContext:       baseCtx,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// HTTPS-only path (default with --tls): one port, no redirect trap.
	if redirectHTTPAddr == "" {
		log.Printf("listening on https://%s (no HTTP redirect listener)", httpsLn.Addr())
		g, gctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			err := httpsSrv.ServeTLS(httpsLn, certFile, keyFile)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("https: %w", err)
			}
			return nil
		})
		g.Go(func() error {
			<-gctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()
			return httpsSrv.Shutdown(shutdownCtx)
		})
		return g.Wait()
	}

	httpLn, err := net.Listen("tcp", redirectHTTPAddr)
	if err != nil {
		if cerr := httpsLn.Close(); cerr != nil {
			return fmt.Errorf("listen http %s: %w (also close https: %v)", redirectHTTPAddr, err, cerr)
		}
		return fmt.Errorf("listen http %s: %w", redirectHTTPAddr, err)
	}

	httpSrv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}
			targetHost := host
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
	log.Printf("listening on http://%s (redirect → HTTPS; ensure HTTPS port is reachable)", httpLn.Addr())

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
