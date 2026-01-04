package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/docmgr/internal/httpapi"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/spf13/cobra"
)

func newServeCommand() *cobra.Command {
	var (
		addr       string
		root       string
		corsOrigin string
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the docmgr HTTP API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			root = workspace.ResolveRoot(root)
			mgr := httpapi.NewIndexManager(root)

			if _, err := mgr.Refresh(ctx); err != nil {
				return fmt.Errorf("failed to build index on startup: %w", err)
			}

			srv := &http.Server{
				Addr:    addr,
				Handler: httpapi.NewServer(mgr, httpapi.ServerOptions{CORSOrigin: corsOrigin}).Handler(),
			}

			go func() {
				<-ctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = srv.Shutdown(shutdownCtx)
			}()

			err := srv.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return err
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8787", "Bind address for the HTTP server")
	cmd.Flags().StringVar(&root, "root", "ttmp", "Docs root directory")
	cmd.Flags().StringVar(&corsOrigin, "cors-origin", "", "If set, add CORS headers for this origin (for browser-based UIs)")

	return cmd
}
