package deploy

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient/helmfake"
)

func TestRollbackRelease(t *testing.T) {
	const dep = "demo-dev"
	managed := map[string]string{helmclient.LabelDeployment: dep}
	cases := []struct {
		name      string
		status    string
		labels    map[string]string
		prev      int
		wantCalls []string
		wantErr   bool
	}{
		{"deployed rolls back", "deployed", managed, 2, []string{"rollback ns/app to 2"}, false},
		{"pending upgrade rolls back", "pending-upgrade", managed, 1, []string{"rollback ns/app to 1"}, false},
		{"failed with history rolls back", "failed", managed, 1, []string{"rollback ns/app to 1"}, false},
		{"pending install clears record", "pending-install", managed, 0, []string{"uninstall ns/app"}, false},
		{"no deployed history clears record", "failed", managed, 0, []string{"uninstall ns/app"}, false},
		{"other deployment refused", "deployed", map[string]string{helmclient.LabelDeployment: "other"}, 1, nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hc := helmfake.New(helmfake.Release{Name: "app", Namespace: "ns", Status: c.status, Revision: 3, Labels: c.labels})
			if c.prev > 0 {
				hc.PreviousRevision["ns/app"] = c.prev
			}
			o := RunOptions{FilePath: "releases.yaml", Stdout: &bytes.Buffer{}}
			err := rollbackRelease(context.Background(), o, hc, dep, "app", "ns")
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if !reflect.DeepEqual(hc.Calls, c.wantCalls) {
				t.Fatalf("calls %q, want %q", hc.Calls, c.wantCalls)
			}
		})
	}
}
