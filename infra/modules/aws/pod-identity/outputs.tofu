output "pod_identity_role_arns" {
  value = { for k, r in aws_iam_role.entry : k => r.arn }
}