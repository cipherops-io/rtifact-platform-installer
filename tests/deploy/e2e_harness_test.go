//go:build e2e

package deploy_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy"
)

var safeTestName = regexp.MustCompile(`[^a-z0-9-]+`)

func deploymentPrefix(t *testing.T) string {
	t.Helper()
	name := strings.ToLower(t.Name())
	name = safeTestName.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if len(name) > 36 {
		name = name[:36]
	}
	return "e2e-" + name
}

func isolateFixture(t *testing.T, doc *deploy.Document) {
	t.Helper()
	p := deploymentPrefix(t)
	doc.NamePrefix = p
	repl := map[string]string{
		"deploy-a": p + "-ns-a",
		"deploy-b": p + "-ns-b",
	}
	for i := range doc.Namespaces {
		if n, ok := repl[doc.Namespaces[i].Name]; ok {
			doc.Namespaces[i].Name = n
		}
	}
	for i := range doc.Releases {
		if n, ok := repl[doc.Releases[i].Namespace]; ok {
			doc.Releases[i].Namespace = n
		}
	}
}

func e2eCleanupDeploy(t *testing.T, file string, o deploy.RunOptions) {
	t.Helper()
	t.Cleanup(func() {
		if err := deploy.Destroy(context.Background(), o); err != nil {
			t.Logf("cleanup destroy: %v", err)
		}
	})
}