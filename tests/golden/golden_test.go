package golden

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "regenerate golden files under tests/golden/<example>/")

var examples = []string{
	"examples/azure/minimal.yaml",
	"examples/aws/minimal.yaml",
	"examples/aws/aws-sbx2.yaml",
	"examples/gcp/minimal.yaml",
	"examples/ai-os-centralindia/config.yaml",
}

func exampleSlug(rel string) string {
	base := filepath.Base(rel)
	dir := filepath.Base(filepath.Dir(rel))
	if dir == "examples" {
		return base[:len(base)-len(".yaml")]
	}
	return dir + "-" + base[:len(base)-len(".yaml")]
}

// goldenHome is fixed so backend local paths and artifact dirs are stable across runs.
const goldenHome = "/tmp/rtifact-golden-home"

func TestGoldenExamples(t *testing.T) {
	root := initContentRoot(t)
	t.Setenv("HOME", goldenHome)

	for _, rel := range examples {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			goldenDir := filepath.Join("tests", "golden", exampleSlug(rel))
			snap, err := BuildSnapshot(root, rel)
			if err != nil {
				t.Fatal(err)
			}
			if *update {
				dir := filepath.Join(root, goldenDir)
				if err := snap.WriteDir(dir); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated %s", goldenDir)
				return
			}
			want, err := LoadSnapshot(filepath.Join(root, goldenDir))
			if err != nil {
				t.Fatalf("load golden %s: %v (run go test ./tests/golden/... -update)", goldenDir, err)
			}
			assertJSONEqual(t, "merged", snap.Merged, want.Merged)
			assertJSONEqual(t, "validation", snap.Validation, want.Validation)
			assertJSONEqual(t, "backends", snap.Backends, want.Backends)
			assertJSONEqual(t, "layers", snap.Layers, want.Layers)
		})
	}
}

func assertJSONEqual(t *testing.T, name string, got, want any) {
	t.Helper()
	gb, err := marshalSortedJSON(got)
	if err != nil {
		t.Fatal(err)
	}
	wb, err := marshalSortedJSON(want)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gb, wb) {
		t.Fatalf("%s mismatch\n--- want\n--- got\n(diff not shown; use -update only in W0)", name)
		t.Logf("want:\n%s", wb)
		t.Logf("got:\n%s", gb)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}