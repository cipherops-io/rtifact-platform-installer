package outputs

import (
	"fmt"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/maputil"
)

// awsKubeconfigYAML builds an EKS kubeconfig using the aws eks get-token exec plugin.
func awsKubeconfigYAML(flat map[string]any, region string) string {
	name := maputil.FirstString(flat, "cluster_name", "aks_name")
	endpoint := asString(flat["cluster_endpoint"])
	ca := asString(flat["cluster_certificate_authority_data"])
	if name == "" || endpoint == "" || ca == "" || region == "" {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "apiVersion: v1\nkind: Config\n")
	fmt.Fprintf(&b, "clusters:\n- cluster:\n    certificate-authority-data: %s\n    server: %s\n  name: %s\n", ca, endpoint, name)
	fmt.Fprintf(&b, "contexts:\n- context:\n    cluster: %s\n    user: %s\n  name: %s\n", name, name, name)
	fmt.Fprintf(&b, "current-context: %s\n", name)
	fmt.Fprintf(&b, "users:\n- name: %s\n  user:\n    exec:\n      apiVersion: client.authentication.k8s.io/v1beta1\n      command: aws\n      args:\n      - eks\n      - get-token\n      - --cluster-name\n      - %s\n      - --region\n      - %s\n", name, name, region)
	return b.String()
}
