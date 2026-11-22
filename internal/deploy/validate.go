package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	"gopkg.in/yaml.v3"
)

var (
	chartNamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?$`)
)

// ValidateFile runs JSON Schema plus semantic checks for a deploy config.
func ValidateFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return usererr.WithHint("Cannot read the deploy file.", "Check the path.", err.Error())
	}
	var data any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return usererr.Wrap(err, "Deploy file is not valid YAML.", "Fix indentation and quoting.")
	}
	data = maputil.NormalizeYAML(data)
	sch, err := compileSchema()
	if err != nil {
		return err
	}
	if err := sch.Validate(data); err != nil {
		return usererr.Wrap(err, "Deploy file failed JSON schema validation.", "Compare with examples/deploy/minimal.yaml and schemas/deploy.schema.json.")
	}
	doc, err := LoadFile(path)
	if err != nil {
		return err
	}
	if doc.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version must be %d", SchemaVersion)
	}
	if doc.Kind != Kind {
		return fmt.Errorf("kind must be %q", Kind)
	}
	if err := validateSemantic(path, doc); err != nil {
		return err
	}
	return nil
}

func validateSemantic(deployPath string, doc *Document) error {
	if dep := DeploymentID(doc); len(dep) > 63 {
		return usererr.WithHint(
			fmt.Sprintf("Deployment name %q is too long for Kubernetes labels (63 character limit).", dep),
			"Shorten customer.code or environment.",
			fmt.Sprintf("length %d", len(dep)),
		)
	}
	if doc.Defaults.Timeout != "" {
		if _, err := time.ParseDuration(doc.Defaults.Timeout); err != nil {
			return fmt.Errorf("defaults.timeout %q: %w", doc.Defaults.Timeout, err)
		}
	}
	if err := validateRegistryAuth(doc); err != nil {
		return err
	}
	nsSet := make(map[string]struct{}, len(doc.Namespaces))
	for _, n := range doc.Namespaces {
		if _, dup := nsSet[n.Name]; dup {
			return fmt.Errorf("duplicate namespace %q", n.Name)
		}
		nsSet[n.Name] = struct{}{}
		if n.AttachPullSecretToDefaultSA && !n.PullSecret {
			return usererr.WithHint(
				fmt.Sprintf("Namespace %q has attach_pull_secret_to_default_sa without pull_secret.", n.Name),
				"Set pull_secret: true on that namespace, or disable attach_pull_secret_to_default_sa.",
				"",
			)
		}
	}
	names := make(map[string]struct{}, len(doc.Releases))
	dir := filepath.Dir(deployPath)
	for _, rel := range doc.Releases {
		if _, ok := nsSet[rel.Namespace]; !ok {
			return usererr.WithHint(
				fmt.Sprintf("Release %q uses namespace %q, which is not listed under namespaces.", rel.Name, rel.Namespace),
				"Add the namespace under namespaces:, or fix release.namespace.",
				"",
			)
		}
		if _, dup := names[rel.Name]; dup {
			return fmt.Errorf("duplicate release name %q", rel.Name)
		}
		names[rel.Name] = struct{}{}
		if err := validateChartRef(rel.Chart); err != nil {
			return fmt.Errorf("release %q: %w", rel.Name, err)
		}
		if err := validateReleaseVersion(rel.Version); err != nil {
			return fmt.Errorf("release %q: %w", rel.Name, err)
		}
		for _, vf := range rel.ValuesFiles {
			p := vf
			if !filepath.IsAbs(p) {
				p = filepath.Join(dir, vf)
			}
			if _, err := os.Stat(p); err != nil {
				return usererr.WithHint(
					fmt.Sprintf("Release %q: values file not found: %s", rel.Name, vf),
					"Paths are relative to the deploy file directory.",
					err.Error(),
				)
			}
		}
		if err := validatePlaceholdersTree(rel.Values); err != nil {
			return fmt.Errorf("release %q values: %w", rel.Name, err)
		}
		for _, vf := range rel.ValuesFiles {
			p := vf
			if !filepath.IsAbs(p) {
				p = filepath.Join(dir, vf)
			}
			if err := validatePlaceholdersInFile(p); err != nil {
				return fmt.Errorf("release %q values file %q: %w", rel.Name, vf, err)
			}
		}
	}
	for _, rel := range doc.Releases {
		for _, req := range rel.Requires {
			if _, ok := names[req]; !ok {
				return usererr.WithHint(
					fmt.Sprintf("Release %q requires unknown release %q.", rel.Name, req),
					"List required releases in releases: with matching name fields.",
					"",
				)
			}
		}
	}
	order, err := InstallOrder(doc.Releases)
	if err != nil {
		return err
	}
	if len(order) != len(doc.Releases) {
		return fmt.Errorf("release graph: incomplete ordering")
	}
	return nil
}

func validateRegistryAuth(doc *Document) error {
	needsAuth := false
	for _, n := range doc.Namespaces {
		if n.PullSecret {
			needsAuth = true
			break
		}
	}
	if !needsAuth {
		return nil
	}
	if doc.Registry.Auth == nil {
		return usererr.WithHint(
			"A namespace sets pull_secret: true but registry.auth is missing.",
			"Add registry.auth with username_env and password_env.",
			"",
		)
	}
	if doc.Registry.Auth.UsernameEnv == "" || doc.Registry.Auth.PasswordEnv == "" {
		return fmt.Errorf("registry.auth requires username_env and password_env when pull_secret is enabled")
	}
	return nil
}

func validateChartRef(chart string) error {
	if chart == "" {
		return fmt.Errorf("chart is required")
	}
	if strings.HasPrefix(chart, "oci://") {
		rest := strings.TrimPrefix(chart, "oci://")
		if rest == "" || !strings.Contains(rest, "/") {
			return fmt.Errorf("chart %q: invalid oci:// reference (expected oci://host/path/chart)", chart)
		}
		return nil
	}
	if !chartNamePattern.MatchString(chart) {
		return fmt.Errorf("chart %q: must be a chart name under registry.charts_path or an oci:// reference", chart)
	}
	return nil
}

func validateReleaseVersion(version string) error {
	if _, err := semver.NewVersion(version); err != nil {
		return fmt.Errorf("version %q: exact semver required (%v)", version, err)
	}
	return nil
}