
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_event_rule.instance_state_change](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_rule.rebalance](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_rule.scheduled_change](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_rule.spot_interruption](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_target.instance_state_change](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/cloudwatch_event_target) | resource |
| [aws_cloudwatch_event_target.rebalance](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/cloudwatch_event_target) | resource |
| [aws_cloudwatch_event_target.scheduled_change](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/cloudwatch_event_target) | resource |
| [aws_cloudwatch_event_target.spot_interruption](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/cloudwatch_event_target) | resource |
| [aws_eks_access_entry.karpenter_node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_access_entry) | resource |
| [aws_eks_pod_identity_association.karpenter](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eks_pod_identity_association) | resource |
| [aws_iam_instance_profile.karpenter_node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_instance_profile) | resource |
| [aws_iam_role.karpenter_controller](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role.karpenter_node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy.karpenter_controller](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_iam_role_policy_attachment.karpenter_node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_sqs_queue.karpenter](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/sqs_queue) | resource |
| [aws_sqs_queue_policy.karpenter](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/sqs_queue_policy) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| cluster_name | n/a | `string` | n/a | yes |
| karpenter | kubernetes_cluster.karpenter block from the customer config; optional keys fall back to the defaults in main.tofu. | `any` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| interruption_queue_name | n/a |
| node_role_arn | n/a |
<!-- END_TF_DOCS -->
