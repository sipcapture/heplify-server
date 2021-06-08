package input

import (
	"testing"
	"crypto/tls"
	"github.com/stretchr/testify/assert"
)

func TestMinVersionConfig_Valid(t *testing.T) {
	tests := []struct {
		input string
		output uint16
	}{
		{"1.0",tls.VersionTLS10},
		{"1.1",tls.VersionTLS11},
		{"1.2",tls.VersionTLS12},
		{"1.3",tls.VersionTLS13},
	}
	for _,test := range(tests) {
		func(t *testing.T,input string, output uint16) {
			minVersion := parseTLSVersion(input)
			assert.Equal(t, minVersion ,output)
		}(t,test.input,test.output)
	}
}

func TestMinVersionConfig_Invalid(t *testing.T) {
	tests := []struct {
		input string
		output uint16
	}{
		{"",tls.VersionTLS12},
		{"10",tls.VersionTLS12},
		{"11",tls.VersionTLS12},
		{"12",tls.VersionTLS12},
		{"13",tls.VersionTLS12},
		{"a",tls.VersionTLS12},
		{"A",tls.VersionTLS12},
		{"TLS1.0",tls.VersionTLS12},
		{"TLS1.1",tls.VersionTLS12},
		{"TLS1.2",tls.VersionTLS12},
		{"TLS1.3",tls.VersionTLS12},
	}
	for _,test := range(tests) {
		func(t *testing.T,input string, output uint16) {
			minVersion := parseTLSVersion(input)
			assert.Equal(t, minVersion ,output)
		}(t,test.input,test.output)
	}
}
