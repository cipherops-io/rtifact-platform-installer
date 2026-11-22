package archtest

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const module = "github.com/cipherops-io/rtifact-platform-installer"

// Packages that must not import internal/cloud/<provider> (only cloud + registry).
var providerImportGuard = []string{
	"internal/runner",
	"internal/config",
	"internal/cli",
	"internal/bootstrap",
	"internal/preflight",
	"internal/outputs",
	"internal/quota",
	"internal/deploy",
}

// Non-test files allowed to compare cloud id strings (customer contract / defaults).
var cloudStringAllowlist = map[string]bool{
	"internal/config/merge.go":        true,
	"internal/config/init_config.go": true,
	"internal/cli/root.go":            true,
	"internal/outputs/outputs.go":     true,
	"internal/quota/from_config.go":   true,
}

func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

func goFilesUnder(root, relDir string, includeTests bool) ([]string, error) {
	base := filepath.Join(root, relDir)
	var out []string
	err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "testdata" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if !includeTests && strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}
		out = append(out, path)
		return nil
	})
	return out, err
}

func importPaths(goFile string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, goFile, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, imp := range f.Imports {
		p := strings.Trim(imp.Path.Value, `"`)
		paths = append(paths, p)
	}
	return paths, nil
}

func isForbiddenProviderImport(importPath string) bool {
	if !strings.HasPrefix(importPath, module+"/internal/cloud/") {
		return false
	}
	if importPath == module+"/internal/cloud" ||
		importPath == module+"/internal/cloud/registry" {
		return false
	}
	return true
}

func cloudStringViolations(goFile, root string) ([]string, error) {
	rel, err := filepath.Rel(root, goFile)
	if err != nil {
		return nil, err
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "internal/cloud/") {
		return nil, nil
	}
	if strings.HasSuffix(rel, "_test.go") {
		return nil, nil
	}
	if cloudStringAllowlist[rel] {
		return nil, nil
	}
	if !strings.HasPrefix(rel, "internal/") {
		return nil, nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, goFile, nil, 0)
	if err != nil {
		return nil, err
	}

	const cloudName = "azure"
	const cloudName2 = "aws"
	const cloudName3 = "gcp"
	var viol []string

	ast.Inspect(f, func(n ast.Node) bool {
		be, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}
		if be.Op != token.EQL && be.Op != token.NEQ {
			return true
		}
		for _, lit := range []struct {
			x, y ast.Expr
		}{
			{be.X, be.Y},
			{be.Y, be.X},
		} {
			if s, ok := lit.x.(*ast.BasicLit); ok && s.Kind == token.STRING {
				v := strings.Trim(s.Value, `"`)
				if v == cloudName || v == cloudName2 || v == cloudName3 {
					pos := fset.Position(be.Pos())
					viol = append(viol, rel+":"+itoa(pos.Line)+": cloud string compare")
				}
			}
		}
		return true
	})

	// switch on cloud name literals
	ast.Inspect(f, func(n ast.Node) bool {
		clause, ok := n.(*ast.CaseClause)
		if !ok {
			return true
		}
		for _, expr := range clause.List {
			if s, ok := expr.(*ast.BasicLit); ok && s.Kind == token.STRING {
				v := strings.Trim(s.Value, `"`)
				if v == cloudName || v == cloudName2 || v == cloudName3 {
					pos := fset.Position(clause.Pos())
					viol = append(viol, rel+":"+itoa(pos.Line)+": switch on cloud id")
				}
			}
		}
		return true
	})

	return viol, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}