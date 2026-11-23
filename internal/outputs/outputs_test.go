package outputs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

func TestMaterializeCWD(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Chdir(dir)
	flat := map[string]any{
		"kubeconfig_raw":              "apiVersion: v1\nkind: Config\n",
		"bastion_ssh_private_key_pem": "-----BEGIN RSA PRIVATE KEY-----\nMIIB\n-----END RSA PRIVATE KEY-----\n",
	}
	written, err := MaterializeCWD("cm-test", flat, MaterializeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(written) < 2 {
		t.Fatalf("expected at least cwd kubeconfig and bastion, got %v", written)
	}
	for _, name := range []string{"cm-test-kubeconfig.yaml", "cm-test-bastion.pem"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestMaterializeFromArtifacts_missing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got, err := MaterializeFromArtifacts("no-such-prefix-xyz", MaterializeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected nil slice, got %v", got)
	}
	_ = stacks.LayerAppPlatform
}

func writeLayer(t *testing.T, prefix string, layer stacks.Layer, flat map[string]any) {
	t.Helper()
	b, err := json.Marshal(flat)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.ArtifactsDir(prefix), string(layer)+"-outputs.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestMaterializeFromArtifacts_coreNetworkBastion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cwd := t.TempDir()
	t.Chdir(cwd)
	writeLayer(t, "aws-x", stacks.LayerCoreNetwork, map[string]any{
		"bastion_ssh_private_key_pem": "-----BEGIN RSA PRIVATE KEY-----\nMIIB\n-----END RSA PRIVATE KEY-----",
		"bastion_ssh_command":         "ssh -i aws-x-bastion.pem ec2-user@203.0.113.10",
	})
	writeLayer(t, "aws-x", stacks.LayerAppPlatform, map[string]any{
		"kubeconfig_raw":              "apiVersion: v1\nkind: Config\n",
		"bastion_ssh_private_key_pem": nil,
	})

	written, err := MaterializeFromArtifacts("aws-x", MaterializeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	pem := filepath.Join(cwd, "aws-x-bastion.pem")
	if !slices.Contains(written, pem) {
		t.Fatalf("bastion pem not written: %v", written)
	}
	st, err := os.Stat(pem)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("pem mode %v, want 0600", st.Mode().Perm())
	}
	b, _ := os.ReadFile(pem)
	if !strings.HasSuffix(string(b), "-----END RSA PRIVATE KEY-----\n") {
		t.Fatalf("pem missing trailing newline: %q", b)
	}
	if !slices.Contains(written, filepath.Join(cwd, "aws-x-kubeconfig.yaml")) {
		t.Fatalf("kubeconfig not written: %v", written)
	}

	flat, err := loadLayers("aws-x", MaterializeLayers...)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := BastionSSHCommand("aws-x", cwd, flat), "ssh -i "+pem+" ec2-user@203.0.113.10"; got != want {
		t.Fatalf("ssh command %q, want %q", got, want)
	}
}

func TestBastionSSHCommand(t *testing.T) {
	dir := t.TempDir()
	azure := "az network bastion ssh --name b --resource-group rg --target-resource-id /vm --auth-type ssh-key --username azureuser --ssh-key az-bastion.pem"
	for name, tc := range map[string]struct {
		cmd, want string
		pem       bool
	}{
		"no output":            {cmd: "", want: ""},
		"ssm untouched":        {cmd: "aws ssm start-session --target i-1 --region us-east-1", want: "aws ssm start-session --target i-1 --region us-east-1"},
		"pem missing":          {cmd: "ssh -i az-bastion.pem azureuser@10.0.0.4", want: "ssh -i az-bastion.pem azureuser@10.0.0.4"},
		"ssh absolute":         {cmd: "ssh -i az-bastion.pem azureuser@10.0.0.4", pem: true, want: "ssh -i " + filepath.Join(dir, "az-bastion.pem") + " azureuser@10.0.0.4"},
		"azure bastion tunnel": {cmd: azure, pem: true, want: strings.Replace(azure, " az-bastion.pem", " "+filepath.Join(dir, "az-bastion.pem"), 1)},
	} {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(dir, "az-bastion.pem")
			_ = os.Remove(p)
			if tc.pem {
				if err := os.WriteFile(p, []byte("k\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if got := BastionSSHCommand("az", dir, map[string]any{"bastion_ssh_command": tc.cmd}); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
