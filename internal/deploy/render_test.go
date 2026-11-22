package deploy

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
)

func TestPrintPlanGroupedRemovalsLast(t *testing.T) {
	removal := plan.Row{Name: "old", Namespace: "ns", Action: plan.ActionRemove, FromVersion: "1.0.0"}
	res := plan.Result{
		Rows: []plan.Row{
			{Name: "a", Namespace: "ns", Action: plan.ActionInstall, ToVersion: "1.0.0"},
			removal,
		},
		Removals:  []plan.Row{removal},
		HasChange: true,
	}
	var buf bytes.Buffer
	if err := PrintPlan(&buf, redact.New(), res, false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Removals:") {
		t.Fatalf("missing Removals section: %q", out)
	}
	if strings.Index(out, "install") > strings.Index(out, "Removals:") {
		t.Fatalf("removals should follow installs: %q", out)
	}
	if !strings.Contains(out, "Plan: 1 install, 1 removal") {
		t.Fatalf("missing summary: %q", out)
	}
}

func TestPrintPlanJSON(t *testing.T) {
	res := plan.Result{
		Rows: []plan.Row{
			{Name: "a", Namespace: "ns", Action: plan.ActionUpgrade, FromVersion: "1.0.0", ToVersion: "2.0.0"},
		},
		HasChange: true,
	}
	var buf bytes.Buffer
	if err := PrintPlan(&buf, redact.New(), res, true); err != nil {
		t.Fatal(err)
	}
	var doc planJSONOut
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if !doc.HasChange || doc.Summary.Upgrade != 1 {
		t.Fatalf("got %#v", doc)
	}
}