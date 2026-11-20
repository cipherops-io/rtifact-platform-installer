
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_eks_cluster.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_cluster) | resource |
| [aws_eks_node_group.system](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_node_group) | resource |
| [aws_iam_role.cluster](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role.node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy_attachment.cluster_policy](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.node_cni](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.node_ecr](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.node_worker](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy_attachment) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| availability_zones | n/a | `list(string)` | n/a | yes |
| dns_service_ip | n/a | `string` | n/a | yes |
| kubernetes_version | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| node_pools | n/a | <pre>map(object({<br/>    name       = string<br/>    vm_size    = string<br/>    node_count = number<br/>  }))</pre> | n/a | yes |
| node_subnet_id | n/a | `string` | n/a | yes |
| pod_cidr | n/a | `string` | n/a | yes |
| scope_name | n/a | `string` | n/a | yes |
| service_cidr | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| cluster_certificate_authority_data | n/a |
| cluster_endpoint | n/a |
| cluster_id | n/a |
| cluster_name | n/a |
| oidc_issuer_url | n/a |
<!-- END_TF_DOCS -->
