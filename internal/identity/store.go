package identity

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	identityDir  = "identity"
	certFilename = "node.crt"
	keyFilename  = "node.key"

	dirPerm  os.FileMode = 0700 // owner rwx only
	certPerm os.FileMode = 0644 // readable by owner + group; cert is not secret
	keyPerm  os.FileMode = 0600 // owner read/write only — private key
)

// store handles reading and writing PEM files under <dataDir>/identity/.
type store struct {
	dir      string // absolute path to the identity directory
	certPath string
	keyPath  string
}

// newStore constructs a store rooted at <dataDir>/identity.
func newStore(dataDir string) *store {
	dir := filepath.Join(dataDir, identityDir)
	return &store{
		dir:      dir,
		certPath: filepath.Join(dir, certFilename),
		keyPath:  filepath.Join(dir, keyFilename),
	}
}

// write persists keyPEM to node.key (mode 0600) and certPEM to node.crt (mode 0644) under the identity directory, creating it if necessary.
func (s *store) write(keyPEM, certPEM []byte) error {
	if err := os.MkdirAll(s.dir, dirPerm); err != nil {
		return fmt.Errorf("store: create directory %s: %w", s.dir, err)
	}

	if err := writeFile(s.keyPath, keyPEM, keyPerm); err != nil {
		return fmt.Errorf("store: write key: %w", err)
	}

	if err := writeFile(s.certPath, certPEM, certPerm); err != nil {
		_ = os.Remove(s.keyPath)
		return fmt.Errorf("store: write cert: %w", err)
	}

	return nil
}

// readKey returns the raw PEM bytes of the private key file.
func (s *store) readKey() ([]byte, error) {
	data, err := os.ReadFile(s.keyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("store: key file not found at %s", s.keyPath)
		}
		return nil, fmt.Errorf("store: read key %s: %w", s.keyPath, err)
	}
	return data, nil
}

// readCert returns the raw PEM bytes of the certificate file.
func (s *store) readCert() ([]byte, error) {
	data, err := os.ReadFile(s.certPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("store: cert file not found at %s", s.certPath)
		}
		return nil, fmt.Errorf("store: read cert %s: %w", s.certPath, err)
	}
	return data, nil
}

// writeFile writes data to path atomically-ish: write to a temp file in the half-written PEM file that would confuse the PEM decoder on the next start.
func writeFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ghostmesh-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()

	var writeErr error
	defer func() {
		if writeErr != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if _, writeErr = tmp.Write(data); writeErr != nil {
		_ = tmp.Close()
		return fmt.Errorf("write to temp file: %w", writeErr)
	}

	if writeErr = tmp.Close(); writeErr != nil {
		return fmt.Errorf("close temp file: %w", writeErr)
	}

	// Set permissions before rename so the file is never world-readable at any point (the temp file starts with default umask perms).
	if writeErr = os.Chmod(tmpName, perm); writeErr != nil {
		return fmt.Errorf("chmod temp file: %w", writeErr)
	}

	if writeErr = os.Rename(tmpName, path); writeErr != nil {
		return fmt.Errorf("rename %s → %s: %w", tmpName, path, writeErr)
	}

	return nil
}
