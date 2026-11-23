package ocireg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChartRef(t *testing.T) {
	cases := []struct {
		reg   Registry
		chart string
		want  string
	}{
		{Registry{Host: "r.example.com", ChartsPath: "charts"}, "app", "oci://r.example.com/charts/app"},
		{Registry{Host: "r.example.com/", ChartsPath: "/team/charts/"}, "app", "oci://r.example.com/team/charts/app"},
		{Registry{Host: "ignored"}, "oci://other/x/app", "oci://other/x/app"},
	}
	for _, c := range cases {
		got, err := ChartRef(c.reg, c.chart)
		if err != nil || got != c.want {
			t.Errorf("ChartRef(%+v, %q) = %q, %v; want %q", c.reg, c.chart, got, err, c.want)
		}
	}
}

func TestLoadAuth(t *testing.T) {
	reg := Registry{Host: "r", UsernameEnv: "OCIREG_TEST_USER", PasswordEnv: "OCIREG_TEST_PASS"}
	t.Setenv("OCIREG_TEST_USER", "u")
	t.Setenv("OCIREG_TEST_PASS", "")
	if _, err := LoadAuth(reg); err == nil || !strings.Contains(err.Error(), "authentication failed") {
		t.Fatalf("empty password: want auth error, got %v", err)
	}
	t.Setenv("OCIREG_TEST_PASS", "p")
	c, err := LoadAuth(reg)
	if err != nil || c.Username != "u" || c.Password != "p" {
		t.Fatalf("got %+v, %v", c, err)
	}
	if c, err := LoadAuth(Registry{Host: "anon"}); err != nil || c != (Credentials{}) {
		t.Fatalf("anonymous registry: %+v, %v", c, err)
	}
}

func TestNewClient_caFileErrors(t *testing.T) {
	if _, err := NewClient(Registry{CAFile: filepath.Join(t.TempDir(), "missing.pem")}, Credentials{}); err == nil {
		t.Fatal("missing ca_file should fail")
	}
	bad := filepath.Join(t.TempDir(), "bad.pem")
	if err := os.WriteFile(bad, []byte("not a cert"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(Registry{CAFile: bad}, Credentials{}); err == nil || !strings.Contains(err.Error(), "no certificates") {
		t.Fatalf("invalid PEM: %v", err)
	}
	if _, err := NewClient(Registry{PlainHTTP: true}, Credentials{Username: "u", Password: "p"}); err != nil {
		t.Fatalf("plain client: %v", err)
	}
}

func TestChartDigest_noClientOrVersionIsEmpty(t *testing.T) {
	if d, err := ChartDigest(nil, Registry{Host: "r"}, "app", "1.0.0"); d != "" || err != nil {
		t.Fatalf("nil client: %q, %v", d, err)
	}
}
