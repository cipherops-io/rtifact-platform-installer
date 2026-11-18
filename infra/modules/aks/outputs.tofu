output "cluster_id" {
  value = azurerm_kubernetes_cluster.main.id
}

output "cluster_name" {
  value = azurerm_kubernetes_cluster.main.name
}

output "oidc_issuer_url" {
  value = azurerm_kubernetes_cluster.main.oidc_issuer_url
}