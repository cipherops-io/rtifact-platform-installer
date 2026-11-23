package deploy

const (
	// LabelNamespaceCreatedBy marks namespaces created by the deploy engine.
	LabelNamespaceCreatedBy      = "rtifact.ai/created-by"
	LabelNamespaceCreatedByValue = "deploy"
)

func namespaceCreatedByDeploy(labels map[string]string) bool {
	return labels != nil && labels[LabelNamespaceCreatedBy] == LabelNamespaceCreatedByValue
}