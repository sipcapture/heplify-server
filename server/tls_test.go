package input

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sipcapture/heplify-server/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinVersionConfig_Valid(t *testing.T) {
	tests := []struct {
		input  string
		output uint16
	}{
		{"1.0", tls.VersionTLS10},
		{"1.1", tls.VersionTLS11},
		{"1.2", tls.VersionTLS12},
		{"1.3", tls.VersionTLS13},
	}
	for _, test := range tests {
		func(t *testing.T, input string, output uint16) {
			minVersion := parseTLSVersion(input)
			assert.Equal(t, minVersion, output)
		}(t, test.input, test.output)
	}
}

func TestMinVersionConfig_Invalid(t *testing.T) {
	tests := []struct {
		input  string
		output uint16
	}{
		{"", tls.VersionTLS12},
		{"10", tls.VersionTLS12},
		{"11", tls.VersionTLS12},
		{"12", tls.VersionTLS12},
		{"13", tls.VersionTLS12},
		{"a", tls.VersionTLS12},
		{"A", tls.VersionTLS12},
		{"TLS1.0", tls.VersionTLS12},
		{"TLS1.1", tls.VersionTLS12},
		{"TLS1.2", tls.VersionTLS12},
		{"TLS1.3", tls.VersionTLS12},
	}
	for _, test := range tests {
		func(t *testing.T, input string, output uint16) {
			minVersion := parseTLSVersion(input)
			assert.Equal(t, minVersion, output)
		}(t, test.input, test.output)
	}
}

func withTLSConfig(t *testing.T, fn func()) {
	t.Helper()
	original := config.Setting
	t.Cleanup(func() { config.Setting = original })
	fn()
}

func writeTempCertKey(t *testing.T) (certFile, keyFile, caFile string) {
	t.Helper()
	dir := t.TempDir()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	serverTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "test-server"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTmpl, caTmpl, &serverKey.PublicKey, caKey)
	require.NoError(t, err)

	certFile = filepath.Join(dir, "server.crt")
	keyFile = filepath.Join(dir, "server.key")
	caFile = filepath.Join(dir, "ca.crt")

	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER}), 0o600))
	serverKeyBytes, err := x509.MarshalECPrivateKey(serverKey)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: serverKeyBytes}), 0o600))
	require.NoError(t, os.WriteFile(caFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}), 0o600))
	return certFile, keyFile, caFile
}

func TestLoadTLSConfig_RequiresBothCertAndKey(t *testing.T) {
	withTLSConfig(t, func() {
		config.Setting.TLSCertFile = "only-cert.pem"
		config.Setting.TLSKeyFile = ""
		_, err := loadTLSConfig(tls.VersionTLS12)
		require.Error(t, err)
	})
}

func TestLoadTLSConfig_FromFiles(t *testing.T) {
	withTLSConfig(t, func() {
		certFile, keyFile, _ := writeTempCertKey(t)
		config.Setting.TLSCertFile = certFile
		config.Setting.TLSKeyFile = keyFile
		cfg, err := loadTLSConfig(tls.VersionTLS12)
		require.NoError(t, err)
		require.Len(t, cfg.Certificates, 1)
		assert.Equal(t, uint16(tls.VersionTLS12), cfg.MinVersion)
		assert.Equal(t, tls.NoClientCert, cfg.ClientAuth)
	})
}

func TestLoadTLSConfig_RequireClientCertNeedsCA(t *testing.T) {
	withTLSConfig(t, func() {
		certFile, keyFile, _ := writeTempCertKey(t)
		config.Setting.TLSCertFile = certFile
		config.Setting.TLSKeyFile = keyFile
		config.Setting.TLSRequireClientCert = true
		_, err := loadTLSConfig(tls.VersionTLS12)
		require.Error(t, err)
	})
}

func TestLoadTLSConfig_ClientAuthOptional(t *testing.T) {
	withTLSConfig(t, func() {
		certFile, keyFile, caFile := writeTempCertKey(t)
		config.Setting.TLSCertFile = certFile
		config.Setting.TLSKeyFile = keyFile
		config.Setting.TLSClientCAFile = caFile
		config.Setting.TLSRequireClientCert = false
		cfg, err := loadTLSConfig(tls.VersionTLS12)
		require.NoError(t, err)
		assert.Equal(t, tls.VerifyClientCertIfGiven, cfg.ClientAuth)
		require.NotNil(t, cfg.ClientCAs)
	})
}

func TestLoadTLSConfig_ClientAuthRequired(t *testing.T) {
	withTLSConfig(t, func() {
		certFile, keyFile, caFile := writeTempCertKey(t)
		config.Setting.TLSCertFile = certFile
		config.Setting.TLSKeyFile = keyFile
		config.Setting.TLSClientCAFile = caFile
		config.Setting.TLSRequireClientCert = true
		cfg, err := loadTLSConfig(tls.VersionTLS12)
		require.NoError(t, err)
		assert.Equal(t, tls.RequireAndVerifyClientCert, cfg.ClientAuth)
	})
}
