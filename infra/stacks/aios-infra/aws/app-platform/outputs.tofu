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

output "pod_identity_role_arns" {
  value = length(module.pod_identity) > 0 ? module.pod_identity[0].pod_identity_role_arns : {}
}

output "kubeconfig_raw" {
  description = "kubeconfig for the EKS cluster; authenticates with `aws eks get-token` using the caller's AWS credentials."
  sensitive   = true
  value = length(module.eks) > 0 ? yamlencode({
    apiVersion      = "v1"
    kind            = "Config"
    current-context = module.eks[0].cluster_name
    clusters = [{
      name = module.eks[0].cluster_name
      cluster = {
        server                     = module.eks[0].cluster_endpoint
        certificate-authority-data = module.eks[0].cluster_certificate_authority_data
      }
    }]
    contexts = [{
      name = module.eks[0].cluster_name
      context = {
        cluster = module.eks[0].cluster_name
        user    = module.eks[0].cluster_name
      }
    }]
    users = [{
      name = module.eks[0].cluster_name
      user = {
        exec = {
          apiVersion      = "client.authentication.k8s.io/v1beta1"
          command         = "aws"
          args            = ["eks", "get-token", "--cluster-name", module.eks[0].cluster_name, "--region", var.location, "--output", "json"]
          interactiveMode = "Never"
        }
      }
    }]
  }) : null
}
