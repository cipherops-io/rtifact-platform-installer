package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/packaged"
	"github.com/cipherops-io/rtifact-platform-installer/internal/version"
	"github.com/spf13/cobra"
)

type versionInfo struct {
	Version      string   `json:"version"`
	Commit       string   `json:"commit"`
	Clouds       []string `json:"clouds"`
	Mode         string   `json:"mode"`
	Platform     string   `json:"platform"`
	TofuVersion  string   `json:"tofu_version"`
	DeployEngine string   `json:"deploy_engine"`
}

func buildMode() string {
	if packaged.Bundled() {
		return "bundled"
	}
	return "dev"
}

func currentVersion() versionInfo {
	clouds := version.PackagedCloudsList()
	if clouds == nil {
		clouds = []string{}
	}
	return versionInfo{
		Version:      version.Version,
		Commit:       version.GitCommit,
		Clouds:       clouds,
		Mode:         buildMode(),
		Platform:     version.Platform(),
		TofuVersion:  version.TofuVersion,
		DeployEngine: strings.TrimPrefix(version.DeployEngineLine(), "deploy-engine: "),
	}
}

func writeVersion(w io.Writer, info versionInfo, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	case "text", "":
		tofu := info.TofuVersion
		if tofu == "" {
			tofu = "from PATH"
		}
		fmt.Fprintf(w, "rtifact %s (commit %s) clouds=%s mode=%s platform=%s tofu=%s\n",
			info.Version, info.Commit, strings.Join(info.Clouds, ","), info.Mode, info.Platform, tofu)
		fmt.Fprintf(w, "deploy-engine: %s\n", info.DeployEngine)
		return nil
	default:
		return fmt.Errorf("unknown output format %q (text or json)", format)
	}
}

func newVersionCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version, packaged clouds, build mode and platform",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeVersion(cmd.OutOrStdout(), currentVersion(), format)
		},
	}
	cmd.Flags().StringVarP(&format, "output", "o", "text", "text or json")
	return cmd
}
