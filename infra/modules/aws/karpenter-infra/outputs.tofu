output "interruption_queue_name" {
  value = aws_sqs_queue.karpenter.name
}

output "node_role_arn" {
  value = aws_iam_role.karpenter_node.arn
}