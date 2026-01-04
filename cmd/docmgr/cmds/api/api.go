package api

import "github.com/spf13/cobra"

// Attach registers API server commands under the root.
func Attach(root *cobra.Command) error {
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "HTTP API server",
		Long:  "HTTP API server for docmgr (search and index management).",
	}

	serveCmd := newServeCommand()
	apiCmd.AddCommand(serveCmd)
	root.AddCommand(apiCmd)
	return nil
}
