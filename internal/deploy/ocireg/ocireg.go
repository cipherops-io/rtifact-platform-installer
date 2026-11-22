package ocireg

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	"helm.sh/helm/v4/pkg/registry"
)

// Registry describes OCI chart registry connection settings.
type Registry struct {
	Host       string
	ChartsPath string
	CAFile     string
	PlainHTTP  bool
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
	opts := []registry.ClientOption{}
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if reg.PlainHTTP {
		opts = append(opts, registry.ClientOptHTTPClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
		}))
	}
	client, err := registry.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	if cred.Username != "" || cred.Password != "" {
		host := reg.Host
		if err := client.Login(host, registry.LoginOptBasicAuth(cred.Username, cred.Password)); err != nil {
			return nil, usererr.WithHint(
				fmt.Sprintf("Registry %q authentication failed.", host),
				"Check registry credentials and network reachability.",
				err.Error(),
			)
		}
	}
	_ = reg.CAFile
	return client, nil
}