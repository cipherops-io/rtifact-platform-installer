output "cluster_name" {
  value = length(module.eks) > 0 ? module.eks[0].cluster_name : null
}

output "cluster_endpoint" {
  value = length(module.eks) > 0 ? module.eks[0].cluster_endpoint : null
}

output "cluster_certificate_authority_data" {
  value     = length(module.eks) > 0 ? module.eks[0].cluster_certificate_authority_data : null
  sensitive = true
}

output "karpenter_queue_name" {
  value = length(module.karpenter) > 0 ? module.karpenter[0].interruption_queue_name : null
}

output "oidc_issuer_url" {
  value = length(module.eks) > 0 ? module.eks[0].oidc_issuer_url : null
}