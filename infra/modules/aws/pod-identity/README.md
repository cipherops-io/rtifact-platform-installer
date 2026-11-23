
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_eks_pod_identity_association.entry](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_pod_identity_association) | resource |
| [aws_iam_role.entry](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy.inline](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_iam_role_policy_attachment.managed](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy_attachment) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| cluster_name | n/a | `string` | n/a | yes |
| entries | Pod identity associations keyed by entry name. | <pre>map(object({<br/>    namespace       = string<br/>    service_account = string<br/>    policy_arns     = optional(list(string), [])<br/>    policy_json     = optional(string, "")<br/>  }))</pre> | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| pod_identity_role_arns | n/a |
<!-- END_TF_DOCS -->
