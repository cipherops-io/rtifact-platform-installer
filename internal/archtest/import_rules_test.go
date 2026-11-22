package archtest

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestImportRulesNoProviderPackagesOutsideRegistry(t *testing.T) {
	root := repoRoot()
	for _, pkgDir := range providerImportGuard {
		files, err := goFilesUnder(root, pkgDir, false)
		if err != nil {
			t.Fatalf("%s: %v", pkgDir, err)
		}
		for _, goFile := range files {
			imports, err := importPaths(goFile)
			if err != nil {
				t.Fatalf("%s: parse: %v", goFile, err)
			}
			rel, _ := filepath.Rel(root, goFile)
			for _, imp := range imports {
				if isForbiddenProviderImport(imp) {
					t.Errorf("%s imports forbidden provider package %q", rel, imp)
				}
			}
		}
	}
}

func TestRunnerImportRules(t *testing.T) {
	root := repoRoot()
	files, err := goFilesUnder(root, "internal/runner", false)
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		module + "/internal/cloud":          true,
		module + "/internal/cloud/registry": true,
	}
	for _, goFile := range files {
		imports, err := importPaths(goFile)
		if err != nil {
			t.Fatalf("%s: %v", goFile, err)
		}
		rel, _ := filepath.Rel(root, goFile)
		for _, imp := range imports {
			if !strings.HasPrefix(imp, module+"/internal/cloud") {
				continue
			}
			if !allowed[imp] {
				t.Errorf("%s: runner may only import cloud and cloud/registry, got %q", rel, imp)
			}
		}
	}
}

func TestRegistryMayImportProviders(t *testing.T) {
	root := repoRoot()
	files, err := goFilesUnder(root, "internal/cloud/registry", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no files in internal/cloud/registry")
	}
	found := false
	for _, goFile := range files {
		imports, err := importPaths(goFile)
		if err != nil {
			t.Fatalf("%s: %v", goFile, err)
		}
		for _, imp := range imports {
			if strings.Contains(imp, "/internal/cloud/azure") ||
				strings.Contains(imp, "/internal/cloud/aws") ||
				strings.Contains(imp, "/internal/cloud/gcp") {
				found = true
			}
		}
	}
	if !found {
		t.Error("registry should wire cloud providers")
	}
}