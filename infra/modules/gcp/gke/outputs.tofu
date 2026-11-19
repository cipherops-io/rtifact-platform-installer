output "cluster_id" {
  value = google_container_cluster.main.id
}

output "cluster_name" {
  value = google_container_cluster.main.name
}

output "oidc_issuer_url" {
  value = try(google_container_cluster.main.workload_identity_config[0].workload_pool, "")
}