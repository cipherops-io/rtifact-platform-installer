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

func TestDeployPlanNoHelmOrKubernetesImports(t *testing.T) {
	root := repoRoot()
	files, err := goFilesUnder(root, "internal/deploy/plan", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, goFile := range files {
		imports, err := importPaths(goFile)
		if err != nil {
			t.Fatalf("%s: %v", goFile, err)
		}
		rel, _ := filepath.Rel(root, goFile)
		for _, imp := range imports {
			if strings.Contains(imp, "helm.sh/") || strings.Contains(imp, "k8s.io/") || strings.Contains(imp, "sigs.k8s.io/") {
				t.Errorf("%s: plan must not import cluster/helm packages, got %q", rel, imp)
			}
		}
	}
}

func TestDeployHelmImportsBounded(t *testing.T) {
	root := repoRoot()
	files, err := goFilesUnder(root, "internal/deploy", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, goFile := range files {
		imports, err := importPaths(goFile)
		if err != nil {
			t.Fatalf("%s: %v", goFile, err)
		}
		rel, _ := filepath.Rel(root, goFile)
		rel = filepath.ToSlash(rel)
		for _, imp := range imports {
			if !strings.Contains(imp, "helm.sh/") {
				continue
			}
			if !deployHelmImportAllowed(rel, imp) {
				t.Errorf("%s: unexpected helm.sh import %q", rel, imp)
			}
		}
	}
}

func deployHelmImportAllowed(rel, imp string) bool {
	switch {
	case strings.HasPrefix(rel, "internal/deploy/helmclient/"):
		return true
	case strings.HasPrefix(rel, "internal/deploy/ocireg/"):
		return imp == "helm.sh/helm/v4/pkg/registry"
	case strings.HasPrefix(rel, "internal/deploy/crds/"):
		return strings.HasPrefix(imp, "helm.sh/helm/v4/pkg/chart/")
	case strings.HasPrefix(rel, "internal/deploy/"):
		return imp == "helm.sh/helm/v4/pkg/registry"
	default:
		return false
	}
}

func TestDeployForbiddenImports(t *testing.T) {
	root := repoRoot()
	files, err := goFilesUnder(root, "internal/deploy", false)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{
		module + "/internal/runner",
		module + "/internal/config",
		module + "/internal/cloud",
	}
	for _, goFile := range files {
		imports, err := importPaths(goFile)
		if err != nil {
			t.Fatalf("%s: %v", goFile, err)
		}
		rel, _ := filepath.Rel(root, goFile)
		for _, imp := range imports {
			for _, f := range forbidden {
				if imp == f || strings.HasPrefix(imp, f+"/") {
					t.Errorf("%s imports forbidden %q", rel, imp)
				}
			}
		}
	}
}