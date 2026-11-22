package archtest

import (
	"strings"
	"testing"
)

func TestNoCloudStringComparesOutsideCloudPackage(t *testing.T) {
	root := repoRoot()
	files, err := goFilesUnder(root, "internal", false)
	if err != nil {
		t.Fatal(err)
	}
	var viol []string
	for _, goFile := range files {
		v, err := cloudStringViolations(goFile, root)
		if err != nil {
			t.Fatalf("%s: %v", goFile, err)
		}
		viol = append(viol, v...)
	}
	if len(viol) > 0 {
		t.Errorf("cloud string compares outside internal/cloud:\n%s", strings.Join(viol, "\n"))
	}
}