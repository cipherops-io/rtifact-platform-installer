package outputs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/config"
	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
	"github.com/cipherops-io/rtifact-platform-installer/internal/stacks"
)

// MaterializeOpts supplies cloud context needed to build kubeconfig on some providers.
type MaterializeOpts struct {
	AWSRegion string
}

// File names written under the process working directory after apply.
func kubeconfigFile(prefix string) string { return prefix + "-kubeconfig.yaml" }
func bastionKeyFile(prefix string) string { return prefix + "-bastion.pem" }

// MaterializeLayers lists the layers whose outputs carry kubeconfig or bastion key material
// (the AWS bastion lives in core-network, the Azure bastion VM in app-platform).
var MaterializeLayers = []stacks.Layer{stacks.LayerCoreNetwork, stacks.LayerAppPlatform}

// MaterializeFromArtifacts reads saved layer outputs and writes kubeconfig / bastion PEM into cwd.
func MaterializeFromArtifacts(prefix string, opts MaterializeOpts) ([]string, error) {
	flat, err := loadLayers(prefix, MaterializeLayers...)
	if err != nil || len(flat) == 0 {
		return nil, err
	}
	return materializeCWD(prefix, flat, opts)
}

// loadLayers merges saved outputs of the given layers; later layers win on key clashes.
func loadLayers(prefix string, layers ...stacks.Layer) (map[string]any, error) {
	dir := paths.ArtifactsDir(prefix)
	merged := map[string]any{}
	for _, layer := range layers {
		b, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("%s-outputs.json", layer)))
		if err != nil {
			continue
		}
		var flat map[string]any
		if err := json.Unmarshal(b, &flat); err != nil {
			return nil, fmt.Errorf("%s outputs: %w", layer, err)
		}
		for k, v := range flat {
			if v != nil {
				merged[k] = v
			}
		}
	}
	return merged, nil
}

// BastionSSHCommand returns the bastion_ssh_command output with the key file name replaced by
// the absolute path of the PEM written into dir (when that file exists).
func BastionSSHCommand(prefix, dir string, flat map[string]any) string {
	cmd := asString(flat["bastion_ssh_command"])
	if cmd == "" {
		return ""
	}
	pemPath := filepath.Join(dir, bastionKeyFile(prefix))
	if st, err := os.Stat(pemPath); err == nil && !st.IsDir() {
		cmd = strings.ReplaceAll(cmd, " "+bastionKeyFile(prefix), " "+pemPath)
	}
	return cmd
}

// MaterializeCWD writes sensitive cluster / bastion files into the current working directory.
func MaterializeCWD(prefix string, flat map[string]any, opts MaterializeOpts) ([]string, error) {
	return materializeCWD(prefix, flat, opts)
}

func materializeCWD(prefix string, flat map[string]any, opts MaterializeOpts) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var written []string

	kube := maputil.FirstString(flat,
		"kube_admin_config_raw",
		"kubeconfig_raw",
		"kubeconfig_yaml",
		"kubeconfig",
	)
	if kube == "" {
		kube = awsKubeconfigYAML(flat, opts.AWSRegion)
	}
	if kube != "" {
		p := filepath.Join(cwd, kubeconfigFile(prefix))
		if err := os.WriteFile(p, []byte(kube), 0o600); err != nil {
			return written, err
		}
		written = append(written, p)
		artifactKube := filepath.Join(paths.ArtifactsDir(prefix), kubeconfigFile(prefix))
		if err := os.MkdirAll(filepath.Dir(artifactKube), 0o700); err == nil {
			_ = os.WriteFile(artifactKube, []byte(kube), 0o600)
			written = append(written, artifactKube)
		}
	}

	pem := maputil.FirstString(flat,
		"bastion_ssh_private_key_pem",
		"bastion_private_key_pem",
	)
	if pem != "" {
		p := filepath.Join(cwd, bastionKeyFile(prefix))
		if !strings.HasSuffix(pem, "\n") {
			pem += "\n"
		}
		if err := os.WriteFile(p, []byte(pem), 0o600); err != nil {
			return written, err
		}
		written = append(written, p)
	}

	return written, nil
}

// Print shows the most important outputs and local artifact paths for a deployment.
func Print(configPath string) error {
	cfg, err := config.LoadCustomer(configPath)
	if err != nil {
		return err
	}
	prefix := cfg.NamePrefix()
	dir := paths.ArtifactsDir(prefix)
	cwd, _ := os.Getwd()

	_, _ = MaterializeFromArtifacts(prefix, MaterializeOpts{AWSRegion: cfg.Account.Region})

	type row struct {
		key string
		val string
	}
	var rows []row

	addFromLayer := func(layer stacks.Layer, keys ...string) {
		path := filepath.Join(dir, fmt.Sprintf("%s-outputs.json", layer))
		b, err := os.ReadFile(path)
		if err != nil {
			return
		}
		var flat map[string]any
		if json.Unmarshal(b, &flat) != nil {
			return
		}
		for _, k := range keys {
			if v := asString(flat[k]); v != "" {
				rows = append(rows, row{fmt.Sprintf("%s.%s", layer, k), v})
			}
		}
	}

	switch cfg.Cloud {
	case "aws":
		addFromLayer(stacks.LayerCoreNetwork, "vpc_id", "subnet_ids_by_tier", "bastion_public_ip", "bastion_private_ip")
		addFromLayer(stacks.LayerNetworkEdge, "dns_zone_id", "certificate_regional_arn")
	default:
		addFromLayer(stacks.LayerCoreNetwork, "resource_group_name", "vnet_id", "azure_bastion_name")
	}
	addFromLayer(stacks.LayerDataPlatform,
		"key_vault_uri", "acr_id", "postgresql_id", "storage_account_id")
	switch cfg.Cloud {
	case "aws":
		addFromLayer(stacks.LayerAppPlatform, "cluster_name", "cluster_endpoint", "oidc_issuer_url")
	default:
		addFromLayer(stacks.LayerAppPlatform,
			"cluster_name", "aks_name", "bastion_public_ip", "bastion_private_ip", "oidc_issuer_url")
	}

	kubePath := filepath.Join(cwd, kubeconfigFile(prefix))
	if st, err := os.Stat(kubePath); err == nil && !st.IsDir() {
		rows = append(rows, row{"kubeconfig (local)", kubePath})
	}
	pemPath := filepath.Join(cwd, bastionKeyFile(prefix))
	if st, err := os.Stat(pemPath); err == nil && !st.IsDir() {
		rows = append(rows, row{"bastion SSH key (local)", pemPath})
	}
	if flat, err := loadLayers(prefix, MaterializeLayers...); err == nil {
		if cmd := BastionSSHCommand(prefix, cwd, flat); cmd != "" {
			rows = append(rows, row{"bastion SSH command", cmd})
		}
	}

	if len(rows) == 0 {
		fmt.Fprintf(os.Stdout, "No saved outputs under %s\n", dir)
		fmt.Fprintf(os.Stdout, "Run apply first (e.g. rtifact apply %s --stack app-platform).\n", configPath)
		return nil
	}

	fmt.Fprintf(os.Stdout, "Deployment outputs for %q (artifacts: %s)\n\n", prefix, dir)
	for _, r := range rows {
		fmt.Fprintf(os.Stdout, "%-28s %s\n", r.key+":", r.val)
	}
	return nil
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case float64, float32, int, int64, bool:
		return fmt.Sprint(v)
	default:
		return ""
	}
}
