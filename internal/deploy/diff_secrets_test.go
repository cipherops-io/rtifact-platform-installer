package deploy

import (
	"strings"
	"testing"
)

func TestManifestForDiff_masksSecretValuesNotKeys(t *testing.T) {
	live := `apiVersion: v1
kind: Secret
metadata:
  name: creds
stringData:
  user: alice
  pass: secret-live
`
	planned := `apiVersion: v1
kind: Secret
metadata:
  name: creds
stringData:
  user: alice
  pass: secret-planned
  token: new
`
	l, err := manifestForDiff(live)
	if err != nil {
		t.Fatal(err)
	}
	p, err := manifestForDiff(planned)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(l, "secret-live") || strings.Contains(p, "secret-planned") {
		t.Fatalf("secret values leaked:\nlive=%q\nplanned=%q", l, p)
	}
	d := unifiedDiff("live", "planned", l, p)
	if !strings.Contains(d, "token") {
		t.Fatalf("expected key-level token change in diff, got:\n%s", d)
	}
	if strings.Contains(d, "secret-") {
		t.Fatalf("raw secret in diff:\n%s", d)
	}
}