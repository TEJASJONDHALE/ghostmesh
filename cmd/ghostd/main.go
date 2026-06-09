package main

import (
	"fmt"
	"github.com/TEJASJONDHALE/ghostmesh/internal/config"
	//	"github.com/TEJASJONDHALE/ghostmesh/internal/discovery"
	"github.com/TEJASJONDHALE/ghostmesh/internal/identity"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	root := &cobra.Command{
		Use:   "ghostd",
		Short: "GhostMesh daemon - runs on evert mode",
	}

	root.AddCommand(startCmd())
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func startCmd() *cobra.Command {
	cfg := config.Default()

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the GhostMesh daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("[ghostd] starting node=%s gossip_port=%d daemon_port=%d data_dir=%s\n",
				cfg.NodeName, cfg.GossipPort, cfg.DaemonPort, cfg.DataDir)

			if cfg.ClusterToken == "" {
				return fmt.Errorf("--token is required")
			}

			fmt.Println("[ghostd] cluster token accepted")
			fmt.Println("[ghostd] daemon running — press Ctrl+C to stop")

			// Block forever - signal handling
			id, err := identity.LoadOrCreate(cfg)
			if err != nil {
				return fmt.Errorf("identity init failed: %w", err)
			}
			fmt.Printf("[ghostd] identity initialized %s\n", id)

			// ... next: registry.Open(), discovery.Start(), etc.
			select {}
		},
	}

	cmd.Flags().StringVar(&cfg.ClusterToken, "token", "", "Cluster join token (required)")
	cmd.Flags().IntVar(&cfg.GossipPort, "gossip-port", cfg.GossipPort, "Gossip listen port")
	cmd.Flags().IntVar(&cfg.DaemonPort, "port", cfg.DaemonPort, "Daemon listen port")
	cmd.Flags().StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "Directory for certs and registry")
	cmd.Flags().StringVar(&cfg.NodeName, "name", cfg.NodeName, "Node name (defaults to hostname)")

	return cmd
}
