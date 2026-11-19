output "gcs_bucket_name" { value = google_storage_bucket.tfstate.name }
output "resource_group_name" {
  description = "Compatibility alias for installer discovery."
  value       = "${var.name_prefix}-tfstate"
}
