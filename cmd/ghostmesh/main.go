package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "ghostmesh",
		Short: "GhostMesh CLI - control mesh",
	}

	root.AddCommand(exposeCmd())
	root.AddCommand(statusCmd())
	root.AddCommand(peersCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func exposeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "expose <service:port>",
		Short: "Expose a local service to the mesh",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("[ghostmesh] exposing service: %s\n", args[0])
			fmt.Println("[ghostmesh] not yet implemented")
			return nil
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon and mesh status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[ghostmesh] status - not yet implemented")
			return nil
		},
	}
}

func peersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peers",
		Short: "List known peers in the mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[ghostmesh] peers - not yet implemented")
			return nil
		},
	}
}
