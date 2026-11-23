package deploy

import "github.com/cipherops-io/rtifact-platform-installer/internal/deploy/model"

const (
	SchemaVersion = model.SchemaVersion
	Kind          = model.Kind
)

type (
	Document      = model.Document
	Customer      = model.Customer
	Cluster       = model.Cluster
	Registry      = model.Registry
	RegistryAuth  = model.RegistryAuth
	PullSecretRef = model.PullSecretRef
	Defaults      = model.Defaults
	Namespace     = model.Namespace
	Release       = model.Release
)
