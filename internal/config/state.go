package config

import (
	"os"
	"path/filepath"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
	"github.com/cipherops-io/rtifact-platform-installer/internal/paths"
)

// UsesLocalState reports whether OpenTofu state is kept on local disk.
func (c *CustomerConfig) UsesLocalState() bool {
	if c.State != nil && c.State.Backend == "local" {
		return true
	}
	if c.Backend != nil {
		switch c.Backend.Type {
		case "azurerm", "s3", "gcs":
			return false
		}
	}
	if c.State != nil {
		switch c.State.Backend {
		case "azurerm", "s3", "gcs":
			return false
		}
	}
	return true
}

// CloudStateSpec builds the provider-neutral state description for this config.
func CloudStateSpec(c *CustomerConfig, profileID string) cloud.StateSpec {
	spec := cloud.StateSpec{
		Profile:    profileID,
		NamePrefix: c.NamePrefix(),
		Account: cloud.Account{
			SubscriptionID: c.Account.SubscriptionID,
			TenantID:       c.Account.TenantID,
			AccountID:      c.Account.AccountID,
			ProjectID:      c.Account.ProjectID,
			Region:         c.Account.Region,
			Environment:    c.Account.Environment,
			UseMSI:         c.Account.Auth.UseMSI,
		},
		Local:     c.UsesLocalState(),
		AirGapped: c.IsAirGapped(),
	}
	if spec.Local {
		spec.LocalDir = filepath.Join(paths.ArtifactsDir(c.NamePrefix()), "state")
		_ = os.MkdirAll(spec.LocalDir, 0o755)
	}
	if c.State != nil && c.State.BucketName != "" {
		spec.StateBucketName = c.State.BucketName
	}
	if b := c.Backend; b != nil {
		spec.Override = &cloud.BackendOverride{
			ResourceGroupName:  b.ResourceGroupName,
			StorageAccountName: b.StorageAccountName,
			ContainerName:      b.ContainerName,
			Key:                b.Key,
			S3Bucket:           b.S3Bucket,
			DynamoDBTable:      b.DynamoDBTable,
			S3Region:           b.S3Region,
			GCSBucket:          b.GCSBucket,
			GCSPrefix:          b.GCSPrefix,
		}
	}
	return spec
}
