# ADR 0009: Multicloud state, naming, and CLI

## Status

Accepted

## Context

The `aios-infra` profile deploys the same six layers on Azure, AWS, and GCP. Customers pick `cloud:` in YAML; release builds may embed one or all clouds.

## Decision

1. **State backends:** `azurerm` (Azure), `s3` + DynamoDB lock (AWS), `gcs` (GCP), or `local` for bootstrap. Air-gapped verification uses cloud CLIs only (`az`, `aws`, `gcloud`).
2. **Naming:** Azure [CAF abbreviations](https://learn.microsoft.com/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations); AWS [CAF](https://docs.aws.amazon.com/cloud-adoption-framework/) and [resource naming](https://docs.aws.amazon.com/prescriptive-guidance/latest/optimization-aws-resources/naming.html); GCP [Architecture Framework](https://cloud.google.com/architecture/framework) and [naming](https://cloud.google.com/compute/docs/naming-resources).
3. **Layer contract:** `core-network` exposes stable output keys for `terraform_remote_state` (see [layer-contracts.md](../layer-contracts.md)).
4. **CLI:** Binary `rtifact`. Top-level commands read `cloud:` from YAML; `rtifact azure|aws|gcp` groups enforce cloud match. Single-cloud builds register only their group.
5. **Profile:** `aios-infra` replaces `ai-os-k8s`; the latter remains a one-release alias when loading profiles.

## Consequences

- Three parallel module trees under `infra/modules/<cloud>/`.
- `PACKAGED_CLOUDS` build variable filters embedded IaC and registry.