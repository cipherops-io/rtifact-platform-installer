package helmclient

const (
	LabelDeployment   = "rtifact.rtifact.ai/deployment"
	LabelValuesDigest = "rtifact.rtifact.ai/values-digest"
	LabelChart        = "rtifact.rtifact.ai/chart"
)

func labels(deployment, chart, digest string) map[string]string {
	return map[string]string{
		LabelDeployment:   deployment,
		LabelValuesDigest: digest,
		LabelChart:        chart,
	}
}