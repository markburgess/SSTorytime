package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

		publicFS, err := server.PublicFS()
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		mux.Handle("/", http.FileServer(http.FS(publicFS)))

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
