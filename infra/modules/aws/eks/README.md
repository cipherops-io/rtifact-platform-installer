## Karpenter discovery

The cluster **extra** security group and node subnets carry `karpenter.sh/discovery=<cluster_name>`. Karpenter provisions nodes in subnets tagged for discovery; the initial node group uses the same subnets.

<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_eks_access_entry.admin](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_access_entry) | resource |
| [aws_eks_access_entry.node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_access_entry) | resource |
| [aws_eks_access_policy_association.admin](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_access_policy_association) | resource |
| [aws_eks_addon.addon](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_addon) | resource |
| [aws_eks_cluster.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_cluster) | resource |
| [aws_eks_node_group.pool](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_node_group) | resource |
| [aws_iam_role.cluster](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role.node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy_attachment.cluster](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_launch_template.node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/launch_template) | resource |
| [aws_security_group.cluster_extra](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group.service](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/security_group) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| kubernetes_cluster | kubernetes_cluster block from the customer config; optional keys fall back to the defaults in main.tofu. | `any` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| subnet_ids_by_tier | n/a | `map(list(string))` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vpc_cidrs | n/a | `list(string)` | n/a | yes |
| vpc_id | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| cluster_certificate_authority_data | n/a |
| cluster_endpoint | n/a |
| cluster_name | n/a |
| node_role_arn | n/a |
| oidc_issuer_url | n/a |
| service_security_group_ids | n/a |
<!-- END_TF_DOCS -->
