package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

const (
	// certTTL is how long a self-signed node cert is valid.
	certTTL = 10 * 365 * 24 * time.Hour // 10 years

	keyAlgorithm = "ECDSA-P256"
)

// generateCert creates a fresh ECDSA P-256 private key and a self-signed x509 certificate. It returns both as PEM-encoded byte slices.
func generateCert(nodeName string) (keyPEM, certPEM []byte, err error) {
	// 1. Generate ECDSA P-256 key.
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key (%s): %w", keyAlgorithm, err)
	}

	// 2. Build the x509 template.
	serial, err := randomSerial()
	if err != nil {
		return nil, nil, fmt.Errorf("generate serial: %w", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   nodeName,
			Organization: []string{"GhostMesh"},
		},
		NotBefore: now.Add(-time.Minute), // small back-date guards clock skew
		NotAfter:  now.Add(certTTL),

		// Key usage for mutual TLS: both sides need KeyEncipherment + DigitalSignature, and the cert is its own CA for self-signing.
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true, // self-signed — no CA chain

		// SANs: loopback for Phase 1 single-host testing.
		// Phase 2 will add the node's LAN IP here.
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{nodeName, "localhost"},
	}

	// 3. Self-sign: parent == template, pub == priv.Public().
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("sign certificate: %w", err)
	}

	// 4. Encode private key to PKCS#8 PEM.
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key: %w", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyDER,
	})

	// 5. Encode certificate to PEM.
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return keyPEM, certPEM, nil
}

// fingerprintOf returns the hex-encoded SHA-256 fingerprint of the certificate's raw SubjectPublicKeyInfo bytes.
func fingerprintOf(cert *x509.Certificate) string {
	pubDER, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		h := sha256.Sum256(cert.Raw)
		return hex.EncodeToString(h[:])
	}
	h := sha256.Sum256(pubDER)
	return hex.EncodeToString(h[:])
}

// randomSerial generates a cryptographically random 128-bit certificate serial number, satisfying RFC 5280 § 4.1.2.2.
func randomSerial() (*big.Int, error) {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, err
	}
	return serial, nil
}
