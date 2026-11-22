package helmclient

const LabelDeployment = "rtifact.ai/deployment"

func labels(deployment string) map[string]string {
	return map[string]string{LabelDeployment: deployment}
}