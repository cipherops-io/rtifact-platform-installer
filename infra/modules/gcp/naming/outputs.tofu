output "resource_group_name" { value = local.scope_name }
output "resource_group_id" { value = "projects/${var.project_id}" }
output "tags" { value = local.tags }
output "scope_name" { value = var.prefix }
