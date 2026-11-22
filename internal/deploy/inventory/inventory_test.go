package inventory

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestParseEmpty(t *testing.T) {
	got, err := Parse("  ")
	if err != nil || got != nil {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestParseV1Array(t *testing.T) {
	raw := `[{"name":"a","namespace":"ns","chart":"a","version":"1.0.0","values_digest":"d","protect":true}]`
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "a" || got[0].ValuesDigest != "d" || !got[0].Protect || got[0].ChartDigest != "" {
		t.Fatalf("got %#v", got)
	}
}

func TestParseObjectWithoutVersionIsV1(t *testing.T) {
	raw := `{"entries":[{"name":"a","namespace":"ns","chart":"a","version":"1.0.0","values_digest":"d"}]}`
	got, err := Parse(raw)
	if err != nil || len(got) != 1 || got[0].Version != "1.0.0" {
		t.Fatalf("got %#v, %v", got, err)
	}
}

func TestParseV2(t *testing.T) {
	raw := `{"schema_version":2,"entries":[{"name":"a","namespace":"ns","chart":"a","version":"1.0.0",
	  "chart_digest":"sha256:abc","values_digest":"d","requires":["db"],"protect":true,"updated_at":"2026-01-02T03:04:05Z"}]}`
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	e := got[0]
	if e.ChartDigest != "sha256:abc" || len(e.Requires) != 1 || e.Requires[0] != "db" || !e.Protect ||
		!e.UpdatedAt.Equal(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)) {
		t.Fatalf("got %#v", e)
	}
}

func TestParseErrors(t *testing.T) {
	tests := map[string]string{
		"newer version":   `{"schema_version":3,"entries":[]}`,
		"unknown field":   `{"schema_version":2,"entries":[],"extra":1}`,
		"missing name":    `{"schema_version":2,"entries":[{"namespace":"ns"}]}`,
		"duplicate entry": `[{"name":"a","namespace":"ns"},{"name":"a","namespace":"ns"}]`,
		"not json":        `nope`,
	}
	for name, raw := range tests {
		if _, err := Parse(raw); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestMarshalRoundTripSortedAndStamped(t *testing.T) {
	now := time.Date(2026, 5, 6, 7, 8, 9, 0, time.UTC)
	b, err := Marshal([]Entry{
		{Name: "b", Namespace: "ns", Chart: "b", Version: "1.0.0", ValuesDigest: "d"},
		{Name: "a", Namespace: "ns", Chart: "a", Version: "1.0.0", ValuesDigest: "d", Requires: []string{"b"},
			UpdatedAt: now.Add(-time.Hour)},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	var doc Document
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.SchemaVersion != SchemaVersion || doc.Entries[0].Name != "a" || doc.Entries[1].Name != "b" {
		t.Fatalf("doc %#v", doc)
	}
	if !doc.Entries[1].UpdatedAt.Equal(now) || !doc.Entries[0].UpdatedAt.Equal(now.Add(-time.Hour)) {
		t.Fatalf("timestamps %#v", doc.Entries)
	}
	if !strings.Contains(string(b), `"schema_version":2`) {
		t.Fatalf("raw %s", b)
	}
	back, err := Parse(string(b))
	if err != nil || len(back) != 2 {
		t.Fatalf("round trip: %v %v", back, err)
	}
}

func TestMarshalEmptyEntriesIsArray(t *testing.T) {
	b, err := Marshal(nil, time.Now())
	if err != nil || !strings.Contains(string(b), `"entries":[]`) {
		t.Fatalf("%s %v", b, err)
	}
}
