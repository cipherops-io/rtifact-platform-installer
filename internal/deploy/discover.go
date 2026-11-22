package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

// Discovery describes chosen cluster access paths (for plan/apply banners).
type Discovery struct {
	KubeconfigPath string
	KubeconfigFrom string
	OutputsPath    string
	OutputsFrom    string
}

type DiscoverOpts struct {
	FilePath      string
	Doc           *Document
	KubeconfigCLI string
	WorkDir       string
}

// Discover resolves kubeconfig and infra outputs without using ambient KUBECONFIG.
func Discover(opts DiscoverOpts) (Discovery, error) {
	var d Discovery
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}
	prefix := opts.Doc.NamePrefix
	file := opts.FilePath

	// Kubeconfig: flag, file field, cwd, artifact dir.
	if opts.KubeconfigCLI != "" {
		d.KubeconfigPath = opts.KubeconfigCLI
		d.KubeconfigFrom = "--kubeconfig flag"
	} else if opts.Doc.Cluster.Kubeconfig != "" {
		d.KubeconfigPath = opts.Doc.Cluster.Kubeconfig
		d.KubeconfigFrom = "cluster.kubeconfig in file"
	} else {
		local := filepath.Join(opts.WorkDir, prefix+"-kubeconfig.yaml")
		if st, err := os.Stat(local); err == nil && !st.IsDir() {
			d.KubeconfigPath = local
			d.KubeconfigFrom = "discovered in working directory"
		} else {
			artifact := filepath.Join(config.ArtifactsDir(prefix), prefix+"-kubeconfig.yaml")
			if st, err := os.Stat(artifact); err == nil && !st.IsDir() {
				d.KubeconfigPath = artifact
				d.KubeconfigFrom = "discovered in artifact directory"
			}
		}
	}
	if d.KubeconfigPath == "" {
		return d, usererr.WithHint(
			fmt.Sprintf("No kubeconfig for %s.", prefix),
			fmt.Sprintf("Run: rtifact output <infra.yaml>, or set cluster.kubeconfig in %s.", filepath.Base(file)),
			"",
		)
	}
	if !filepath.IsAbs(d.KubeconfigPath) {
		d.KubeconfigPath = filepath.Join(filepath.Dir(file), d.KubeconfigPath)
	}

	// Infra outputs: file field, then artifact dir (optional).
	if opts.Doc.InfraOutputs != "" {
		d.OutputsPath = opts.Doc.InfraOutputs
		d.OutputsFrom = "infra_outputs in file"
	} else {
		artifact := config.ArtifactsDir(prefix)
		if st, err := os.Stat(artifact); err == nil && st.IsDir() {
			d.OutputsPath = artifact
			d.OutputsFrom = "discovered artifact directory"
		}
	}
	if d.OutputsPath != "" && !filepath.IsAbs(d.OutputsPath) {
		d.OutputsPath = filepath.Join(filepath.Dir(file), d.OutputsPath)
	}
	return d, nil
}

// IdentityGuardRefusal is used when kubeconfig server does not match expected (S13).
func IdentityGuardRefusal(expected, got string) error {
	return usererr.WithHint(
		"The kubeconfig points at a different API server than expected.",
		"Fix cluster.kubeconfig, expected_server, or infrastructure outputs before deploying.",
		fmt.Sprintf("expected %q, kubeconfig server %q", expected, got),
	)
}