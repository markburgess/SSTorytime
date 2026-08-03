package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/markburgess/SSTorytime/internal/css"
	"github.com/markburgess/SSTorytime/internal/db"
	"github.com/markburgess/SSTorytime/internal/server"
	"github.com/spf13/cobra"
)

var (
	serveAddr      string
	serveResources string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "HTTP UI and JSON API (TLS deferred — HTTP only for now)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		if _, _, err := db.OpenDSN(ctx, databaseURL); err != nil {
			return fmt.Errorf("database: %w", err)
		}

		if err := css.Ensure(); err != nil {
			return fmt.Errorf("css themes: %w", err)
		}

		publicFS, err := server.PublicFS()
		if err != nil {
			return err
		}

		fileServer := http.FileServer(http.FS(publicFS))
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if strings.HasSuffix(path, ".css") {
				name := strings.TrimSuffix(path, ".css")
				if b, err := css.CSS(name); err == nil {
					w.Header().Set("Content-Type", "text/css; charset=utf-8")
					_, _ = w.Write(b)
					return
				}
			}
			fileServer.ServeHTTP(w, r)
		})

		if serveResources != "" {
			mux.Handle("/Resources/", http.StripPrefix("/Resources/", http.FileServer(http.Dir(serveResources))))
		}

		srv := &http.Server{Addr: serveAddr, Handler: mux}
		go func() {
			log.Printf("listening on http://%s (TLS deferred)", serveAddr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	},
}

func init() {
	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "HTTP listen address")
	serveCmd.Flags().StringVar(&serveResources, "resources", "/mnt", "directory for /Resources/")
}
