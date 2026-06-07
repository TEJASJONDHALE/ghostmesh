package identity

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/TEJASJONDHALE/ghostmesh/internal/config"
)

// Identity holds the node's TLS credentials and its derived node ID.
type Identity struct {
	NodeID string

	cert tls.Certificate

	// leaf is the decoded x509 leaf — kept so callers can inspect fields (SANs, not-after, etc.) without re-parsing on every call.
	leaf *x509.Certificate
}

// LoadOrCreate loads an existing identity from disk, or generates a new one if no identity files are found. It is idempotent — safe to call on every
func LoadOrCreate(cfg *config.Config) (*Identity, error) {
	s := newStore(cfg.DataDir)

	id, err := loadFromStore(s)
	if err == nil {
		return id, nil
	}

	return generateAndStore(s, cfg.NodeName)
}

// GetTLSConfig returns a *tls.Config suitable for use on both the client and server side of a mutual-TLS QUIC connection.
func (id *Identity) GetTLSConfig() *tls.Config {
	return &tls.Config{
		Certificates:       []tls.Certificate{id.cert},
		ClientAuth:         tls.RequireAnyClientCert,
		InsecureSkipVerify: true, //nolint:gosec // intentional — see doc comment
		MinVersion:         tls.VersionTLS13,
	}
}

// String returns a short human-readable summary for log lines.
func (id *Identity) String() string {
	shortID := id.NodeID

	if len(shortID) > 16 {
		shortID = shortID[:16] + "..."
	}

	return fmt.Sprintf(
		"node_id=%s not_after=%s",
		shortID,
		id.leaf.NotAfter.Format("2006-01-02"),
	)
}

// loadFromStore reads key + cert from disk and constructs an Identity.
func loadFromStore(s *store) (*Identity, error) {
	keyPEM, err := s.readKey()
	if err != nil {
		return nil, fmt.Errorf("identity: read key: %w", err)
	}

	certPEM, err := s.readCert()
	if err != nil {
		return nil, fmt.Errorf("identity: read cert: %w", err)
	}

	return assembleIdentity(keyPEM, certPEM)
}

// generateAndStore generates a fresh ECDSA key + self-signed cert, writes them to disk, then returns the assembled Identity.
func generateAndStore(s *store, nodeName string) (*Identity, error) {
	keyPEM, certPEM, err := generateCert(nodeName)
	if err != nil {
		return nil, fmt.Errorf("identity: generate: %w", err)
	}

	if err := s.write(keyPEM, certPEM); err != nil {
		return nil, fmt.Errorf("identity: persist: %w", err)
	}

	return assembleIdentity(keyPEM, certPEM)
}

// assembleIdentity builds an Identity from raw PEM bytes.
func assembleIdentity(keyPEM, certPEM []byte) (*Identity, error) {
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("identity: parse key pair: %w", err)
	}

	leaf, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("identity: parse leaf cert: %w", err)
	}

	nodeID := fingerprintOf(leaf)

	return &Identity{
		NodeID: nodeID,
		cert:   tlsCert,
		leaf:   leaf,
	}, nil
}
