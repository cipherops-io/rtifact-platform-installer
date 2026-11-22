package ocireg

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	"helm.sh/helm/v4/pkg/registry"
)

// Registry describes OCI chart registry connection settings.
type Registry struct {
	Host        string
	ChartsPath  string
	CAFile      string
	PlainHTTP   bool
	UsernameEnv string
	PasswordEnv string
}

// Credentials from environment variable names in the deploy file.
type Credentials struct {
	Username string
	Password string
}

// LoadAuth reads registry auth from the environment.
func LoadAuth(reg Registry) (Credentials, error) {
	var c Credentials
	if reg.UsernameEnv != "" {
		c.Username, _ = os.LookupEnv(reg.UsernameEnv)
	}
	if reg.PasswordEnv != "" {
		var ok bool
		c.Password, ok = os.LookupEnv(reg.PasswordEnv)
		if !ok || c.Password == "" {
			return Credentials{}, usererr.WithHint(
				fmt.Sprintf("Registry %q authentication failed.", reg.Host),
				fmt.Sprintf("Set the environment variable %s before plan or apply.", reg.PasswordEnv),
				fmt.Sprintf("missing or empty %s", reg.PasswordEnv),
			)
		}
	}
	return c, nil
}

// ChartRef builds an oci:// chart reference from registry settings.
func ChartRef(reg Registry, chart string) (string, error) {
	if strings.HasPrefix(chart, "oci://") {
		return chart, nil
	}
	host := strings.TrimSuffix(reg.Host, "/")
	path := strings.Trim(strings.TrimPrefix(reg.ChartsPath, "/"), "/")
	ref := fmt.Sprintf("oci://%s/%s/%s", host, path, chart)
	return ref, nil
}

// NewClient constructs a Helm OCI registry client with registry TLS options.
func NewClient(reg Registry, cred Credentials) (*registry.Client, error) {
	opts := []registry.ClientOption{
		registry.ClientOptWriter(io.Discard),
	}
	if reg.PlainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if reg.CAFile != "" {
		pem, err := os.ReadFile(reg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("registry ca_file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("registry ca_file: no certificates found in %q", reg.CAFile)
		}
		tlsCfg.RootCAs = pool
	}
	opts = append(opts, registry.ClientOptHTTPClient(&http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
	}))
	if cred.Username != "" || cred.Password != "" {
		// Helm's registry.Login persists credentials under ~/.config/helm/registry.
		// ClientOptBasicAuth keeps auth in memory on the registry client authorizer.
		opts = append(opts, registry.ClientOptBasicAuth(cred.Username, cred.Password))
	}
	return registry.NewClient(opts...)
}