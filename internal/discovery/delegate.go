package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/memberlist"

	"github.com/TEJASJONDHALE/ghostmesh/internal/config"
	"github.com/TEJASJONDHALE/ghostmesh/internal/identity"
)

type nodeMeta struct {
	TokenHash string `json:"token_hash"`
	NodeID    string `json:"node_id"`
	NodeName  string `json:"node_name"`
}

type delegate struct {
	meta     []byte
	tokenSum string
}

func newDelegate(cfg *config.Config, id *identity.Identity) *delegate {
	sum := tokenHash(cfg.ClusterToken)

	encoded, err := json.Marshal(nodeMeta{
		TokenHash: sum,
		NodeID:    id.NodeID,
		NodeName:  cfg.NodeName,
	})
	if err != nil {
		panic(fmt.Sprintf("discovery: marshal node meta: %v", err))
	}
	if len(encoded) > memberlist.MetaMaxSize {
		panic(fmt.Sprintf("discovery: node meta too large: %d > %d bytes", len(encoded), memberlist.MetaMaxSize))
	}

	return &delegate{meta: encoded, tokenSum: sum}
}

func (d *delegate) NodeMeta(_ int) []byte             { return d.meta }
func (d *delegate) NotifyMsg(_ []byte)                {}
func (d *delegate) GetBroadcasts(_, _ int) [][]byte   { return nil }
func (d *delegate) LocalState(_ bool) []byte          { return nil }
func (d *delegate) MergeRemoteState(_ []byte, _ bool) {}

func (d *delegate) NotifyJoin(node *memberlist.Node) {
	meta, err := parseMeta(node.Meta)
	if err != nil {
		fmt.Printf("[ghostd] gossip peer_join name=%s addr=%s meta=unparseable err=%v\n",
			node.Name, node.Address(), err)
		return
	}
	if meta.TokenHash != d.tokenSum {
		fmt.Printf("[ghostd] gossip peer_join name=%s addr=%s status=token_mismatch\n",
			node.Name, node.Address())
		return
	}
	fmt.Printf("[ghostd] gossip peer_join name=%s addr=%s node_id=%s token=ok\n",
		node.Name, node.Address(), shortID(meta.NodeID))
}

func (d *delegate) NotifyLeave(node *memberlist.Node) {
	fmt.Printf("[ghostd] gossip peer_leave name=%s addr=%s\n", node.Name, node.Address())
}

func (d *delegate) NotifyUpdate(node *memberlist.Node) {
	meta, err := parseMeta(node.Meta)
	if err != nil {
		return
	}
	fmt.Printf("[ghostd] gossip peer_update name=%s addr=%s node_id=%s\n",
		node.Name, node.Address(), shortID(meta.NodeID))
}

func parseMeta(raw []byte) (*nodeMeta, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty metadata")
	}
	var m nodeMeta
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func shortID(id string) string {
	if len(id) > 16 {
		return id[:16] + "..."
	}
	return id
}

func newDiscardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

// ValidatePeerToken and PeerNodeID are used by the QUIC transport layer in Phase 2.

func ValidatePeerToken(raw []byte, clusterToken string) bool {
	meta, err := parseMeta(raw)
	if err != nil {
		return false
	}
	return meta.TokenHash == tokenHash(clusterToken)
}

func PeerNodeID(raw []byte) string {
	meta, err := parseMeta(raw)
	if err != nil {
		return ""
	}
	return meta.NodeID
}
