package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
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

		// Ensure schema is applied before handlers open sessions.
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

		srv := &http.Server{
			Addr:    serveAddr,
			Handler: mux,
			// Inherit process lifetime so request contexts cancel on SIGINT/SIGTERM.
			BaseContext: func(net.Listener) context.Context { return ctx },
		}
		go func() {
			log.Printf("listening on http://%s (TLS deferred)", serveAddr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}()

		// Wait for process signal (main's NotifyContext) or parent cancel.
		<-ctx.Done()
		// Grace period must outlive the canceled process ctx.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	},
}

func init() {
	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "HTTP listen address")
	serveCmd.Flags().StringVar(&serveResources, "resources", "/mnt", "directory for /Resources/")
}
