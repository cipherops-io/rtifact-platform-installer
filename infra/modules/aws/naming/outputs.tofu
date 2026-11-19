output "resource_group_name" {
  value = aws_resourcegroups_group.main.name
}

output "resource_group_id" {
  value = aws_resourcegroups_group.main.arn
}

output "tags" {
  value = local.tags
}

output "scope_name" {
  value = var.prefix
}
