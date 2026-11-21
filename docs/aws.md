# AWS infrastructure (rtifact installer)

Reference customer: [`examples/aws-sbx2.yaml`](../examples/aws-sbx2.yaml) — second sandbox in account `635702036154`, `ap-south-1`, prefix `rtfc-sbx2`.

## Quick start

```bash
make build-dev
INSTALLER_CONTENT_ROOT=. ./dist/rtifact validate examples/aws-sbx2.yaml
INSTALLER_CONTENT_ROOT=. ./dist/rtifact preflight examples/aws-sbx2.yaml
INSTALLER_CONTENT_ROOT=. ./dist/rtifact bootstrap examples/aws-sbx2.yaml --yes
INSTALLER_CONTENT_ROOT=. ./dist/rtifact apply examples/aws-sbx2.yaml --stack all --accept-preview --yes
```

AWS apply is **preview**: pass `--accept-preview`.

## DNS delegation

When `public_dns.delegated` is `false` and `parent_zone_id` is empty, `network-edge` creates the Route53 zone only. After apply, use stack outputs `dns_name_servers` to add NS records at the parent (`rtifact.ai`). Set `public_dns.delegated: true`, then apply `network-edge` again to issue ACM certificates.

## IAM (operator)

Typical permissions: EC2, VPC, EKS, IAM (roles/policies/instance profiles), S3, DynamoDB (state lock), Route53, ACM, SQS, EventBridge, SSM parameters (AMI), `eks:CreatePodIdentityAssociation`.

## YAML reference

| Section | Layer |
|---------|--------|
| `state` (`bucket_name`, `versioning`, `encryption`) | state-backend |
| `network` (tiered `subnets`, `nat_gateway`, endpoints) | core-network |
| `bastion` | core-network |
| `kubernetes_cluster` (EKS, node groups, addons, Karpenter AWS) | app-platform |
| `public_dns`, `certificates` | network-edge |

Profile overlays: `profiles/aios-infra/defaults-aws.yaml`.