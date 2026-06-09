package discovery

import (
	"fmt"
	"time"

	"github.com/hashicorp/memberlist"

	"github.com/TEJASJONDHALE/ghostmesh/internal/config"
	"github.com/TEJASJONDHALE/ghostmesh/internal/identity"
)

type Gossip struct {
	list     *memberlist.Memberlist
	delegate *delegate
	cfg      *config.Config
}

func Start(cfg *config.Config, id *identity.Identity) (*Gossip, error) {
	d := newDelegate(cfg, id)

	mlCfg, err := buildMemberlistConfig(cfg, d)
	if err != nil {
		return nil, fmt.Errorf("gossip: build config: %w", err)
	}

	list, err := memberlist.Create(mlCfg)
	if err != nil {
		return nil, fmt.Errorf("gossip: create memberlist: %w", err)
	}

	fmt.Printf("[ghostd] gossip started node=%s addr=%s:%d\n",
		cfg.NodeName, mlCfg.BindAddr, mlCfg.BindPort)

	return &Gossip{list: list, delegate: d, cfg: cfg}, nil
}

func (g *Gossip) Join(peers []string) error {
	if len(peers) == 0 {
		return nil
	}

	n, err := g.list.Join(peers)
	if err != nil {
		return fmt.Errorf("gossip: join failed (contacted %d/%d peers): %w", n, len(peers), err)
	}

	fmt.Printf("[ghostd] gossip joined peers_contacted=%d requested=%d\n", n, len(peers))
	return nil
}

func (g *Gossip) Members() []*memberlist.Node {
	all := g.list.Members()
	peers := make([]*memberlist.Node, 0, len(all))
	localName := g.list.LocalNode().Name
	for _, m := range all {
		if m.Name != localName {
			peers = append(peers, m)
		}
	}
	return peers
}

func (g *Gossip) LocalNode() *memberlist.Node {
	return g.list.LocalNode()
}

func (g *Gossip) Stop() error {
	if err := g.list.Leave(3 * time.Second); err != nil {
		fmt.Printf("[ghostd] gossip leave warning: %v\n", err)
	}
	if err := g.list.Shutdown(); err != nil {
		return fmt.Errorf("gossip: shutdown: %w", err)
	}
	fmt.Println("[ghostd] gossip stopped")
	return nil
}

func buildMemberlistConfig(cfg *config.Config, d *delegate) (*memberlist.Config, error) {
	mlCfg := memberlist.DefaultLANConfig()

	mlCfg.Name = cfg.NodeName
	mlCfg.BindAddr = "0.0.0.0"
	mlCfg.BindPort = cfg.GossipPort
	mlCfg.AdvertisePort = cfg.GossipPort
	mlCfg.AdvertiseAddr = "127.0.0.1"
	mlCfg.Delegate = d
	mlCfg.Events = d
	mlCfg.Logger = newDiscardLogger()

	return mlCfg, nil
}
