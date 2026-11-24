package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/assets"
	"github.com/cipherops-io/rtifact-platform-installer/internal/runner"
	"github.com/cipherops-io/rtifact-platform-installer/internal/version"
	"github.com/spf13/cobra"
)

// CLIs reported by doctor; none is required by every command, so a missing one is not a failure.
var doctorTools = []string{"az", "aws", "gcloud", "kubectl", "helm"}

type doctorEnv struct {
	out         io.Writer
	mode        string
	contentRoot string
	extracted   bool
	tofu        string
	tofuErr     error
	mirror      string
	clouds      []string
	wantTofu    string
	platform    string
	lookPath    func(string) (string, error)
}

func newDoctorCmd() *cobra.Command {
	var deep bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check the embedded OpenTofu, provider mirror and local tools",
		Long: "Reports the build mode, embedded OpenTofu, the providers in the mirror for this platform and which cloud CLIs are installed.\n" +
			"--deep also runs `tofu init` for every layer of every packaged cloud against the embedded mirror, without contacting the provider registry.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, extracted, err := assets.ContentRoot()
			if err != nil {
				return fmt.Errorf("no installer content found (not embedded, not in the current directory): %w", err)
			}
			if extracted {
				defer os.RemoveAll(root)
			}
			tofu, mirror, tofuErr := runner.ResolveTofu(root)
			env := doctorEnv{
				out:         cmd.OutOrStdout(),
				mode:        buildMode(),
				contentRoot: root,
				extracted:   extracted,
				tofu:        tofu,
				tofuErr:     tofuErr,
				mirror:      mirror,
				clouds:      version.PackagedCloudsList(),
				wantTofu:    version.TofuVersion,
				platform:    runtime.GOOS + "_" + runtime.GOARCH,
				lookPath:    exec.LookPath,
			}
			return env.run(deep)
		},
	}
	cmd.Flags().BoolVar(&deep, "deep", false, "Run offline tofu init for every layer of every packaged cloud")
	return cmd
}

func (d doctorEnv) run(deep bool) error {
	var gaps []string
	gap := func(format string, a ...any) {
		msg := fmt.Sprintf(format, a...)
		gaps = append(gaps, msg)
		fmt.Fprintf(d.out, "  FAIL %s\n", msg)
	}
	src := d.contentRoot
	if d.extracted {
		src += " (extracted from the binary)"
	}
	fmt.Fprintf(d.out, "mode:         %s\nplatform:     %s\nclouds:       %s\ncontent root: %s\n",
		d.mode, strings.Replace(d.platform, "_", "/", 1), strings.Join(d.clouds, ","), src)

	fmt.Fprintln(d.out, "opentofu:")
	if d.tofuErr != nil {
		gap("%v", d.tofuErr)
	} else if got, err := tofuVersion(d.tofu); err != nil {
		gap("%s does not run: %v", d.tofu, err)
	} else if d.wantTofu != "" && got != d.wantTofu {
		gap("%s is OpenTofu %s, want %s", d.tofu, got, d.wantTofu)
	} else {
		fmt.Fprintf(d.out, "  ok   %s (%s)\n", got, d.tofu)
	}

	fmt.Fprintf(d.out, "providers (%s):\n", d.platform)
	locked, err := lockedProviders(filepath.Join(d.contentRoot, "infra", "stacks"), d.clouds)
	if err != nil {
		gap("%v", err)
	}
	switch {
	case d.mirror == "" && d.mode == "bundled":
		gap("no provider mirror under build/embed/mirror")
	case d.mirror == "":
		fmt.Fprintln(d.out, "  info no provider mirror (dev build): providers come from the registry")
	default:
		have := mirrorProviders(d.mirror, d.platform)
		for _, p := range locked {
			if have[p] {
				fmt.Fprintf(d.out, "  ok   %s\n", p)
			} else {
				gap("%s missing from the mirror", p)
			}
		}
	}

	fmt.Fprintln(d.out, "tools:")
	for _, t := range doctorTools {
		if p, err := d.lookPath(t); err == nil {
			fmt.Fprintf(d.out, "  ok   %s (%s)\n", t, p)
		} else {
			fmt.Fprintf(d.out, "  info %s not installed\n", t)
		}
	}

	if deep {
		fmt.Fprintln(d.out, "offline init:")
		switch {
		case d.tofuErr != nil:
			gap("skipped: no OpenTofu")
		case d.mirror == "":
			gap("skipped: no provider mirror to init from")
		default:
			for _, g := range d.deepInit() {
				gap("%s", g)
			}
		}
	}

	if len(gaps) > 0 {
		return fmt.Errorf("doctor found %d problem(s)", len(gaps))
	}
	fmt.Fprintln(d.out, "doctor: all checks passed")
	return nil
}

func tofuVersion(bin string) (string, error) {
	out, err := exec.Command(bin, "version", "-json").Output()
	if err != nil {
		return "", err
	}
	var v struct {
		Version string `json:"terraform_version"`
	}
	if err := json.Unmarshal(out, &v); err != nil || v.Version == "" {
		return "", fmt.Errorf("unexpected `tofu version -json` output")
	}
	return v.Version, nil
}

var lockProviderRe = regexp.MustCompile(`(?m)^provider "([^"]+)" \{\s*\n\s*version\s*=\s*"([^"]+)"`)

// lockedProviders lists "source version" for every provider pinned by the packaged layers' lock files.
func lockedProviders(stacksRoot string, clouds []string) ([]string, error) {
	set := map[string]bool{}
	for _, c := range clouds {
		locks, _ := filepath.Glob(filepath.Join(stacksRoot, "*", c, "*", ".terraform.lock.hcl"))
		if len(locks) == 0 {
			return nil, fmt.Errorf("no lock files for cloud %s under %s", c, stacksRoot)
		}
		for _, l := range locks {
			b, err := os.ReadFile(l)
			if err != nil {
				return nil, err
			}
			for _, m := range lockProviderRe.FindAllStringSubmatch(string(b), -1) {
				set[m[1]+" "+m[2]] = true
			}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

// mirrorProviders indexes packed mirror archives (<host>/<ns>/<name>/terraform-provider-<name>_<ver>_<platform>.zip).
func mirrorProviders(mirror, platform string) map[string]bool {
	have := map[string]bool{}
	suffix := "_" + platform + ".zip"
	_ = filepath.WalkDir(mirror, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), suffix) {
			return nil
		}
		rel, err := filepath.Rel(mirror, filepath.Dir(p))
		if err != nil {
			return nil
		}
		src := filepath.ToSlash(rel)
		prefix := "terraform-provider-" + filepath.Base(src) + "_"
		if !strings.HasPrefix(d.Name(), prefix) {
			return nil
		}
		have[src+" "+strings.TrimSuffix(strings.TrimPrefix(d.Name(), prefix), suffix)] = true
		return nil
	})
	return have
}

// deepInit copies the packaged layers to a scratch folder and runs tofu init against the mirror only.
func (d doctorEnv) deepInit() []string {
	work, err := os.MkdirTemp("", "rtifact-doctor-")
	if err != nil {
		return []string{err.Error()}
	}
	defer os.RemoveAll(work)
	infra := filepath.Join(work, "infra")
	if err := copyTree(filepath.Join(d.contentRoot, "infra"), infra); err != nil {
		return []string{fmt.Sprintf("copy infra: %v", err)}
	}
	rc := filepath.Join(work, "tofurc")
	if err := runner.WriteProviderRC(rc, d.mirror); err != nil {
		return []string{err.Error()}
	}
	cache := filepath.Join(work, "plugin-cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return []string{err.Error()}
	}
	env := []string{"TF_CLI_CONFIG_FILE=" + rc, "TF_PLUGIN_CACHE_DIR=" + cache, "TF_IN_AUTOMATION=1"}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "TF_CLI_CONFIG_FILE=") && !strings.HasPrefix(kv, "TF_PLUGIN_CACHE_DIR=") && !strings.HasPrefix(kv, "TF_DATA_DIR=") {
			env = append(env, kv)
		}
	}

	var gaps []string
	layers := 0
	for _, c := range d.clouds {
		dirs, _ := filepath.Glob(filepath.Join(infra, "stacks", "*", c, "*", "providers.tofu"))
		if len(dirs) == 0 {
			gaps = append(gaps, fmt.Sprintf("no layers for cloud %s", c))
		}
		for _, p := range dirs {
			layer := filepath.Dir(p)
			name := c + "/" + filepath.Base(layer)
			cmd := exec.Command(d.tofu, "init", "-backend=false", "-input=false", "-no-color")
			cmd.Dir = layer
			cmd.Env = env
			var buf bytes.Buffer
			cmd.Stdout, cmd.Stderr = &buf, &buf
			layers++
			if err := cmd.Run(); err != nil {
				gaps = append(gaps, fmt.Sprintf("%s: tofu init failed: %s", name, lastLines(buf.String(), 6)))
				continue
			}
			fmt.Fprintf(d.out, "  ok   %s\n", name)
		}
	}
	if layers == 0 && len(gaps) == 0 {
		gaps = append(gaps, "no layers to init")
	}
	return gaps
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			if d.Name() == ".terraform" {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, " | ")
}
