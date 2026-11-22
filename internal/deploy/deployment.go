package deploy

// DeploymentID names this deploy file's cluster tenancy (inventory, labels, lock).
func DeploymentID(doc *Document) string {
	return doc.Customer.Code + "-" + doc.Environment
}