data "terraform_remote_state" "core_network" {
  backend = var.upstream_state["core-network"].backend
  config  = var.upstream_state["core-network"].config
}

locals {
  core = data.terraform_remote_state.core_network.outputs
}

module "eks" {
  count = var.features.kubernetes_cluster ? 1 : 0

  source             = "../../../../modules/aws/eks"
  name_prefix        = var.name_prefix
  location           = var.location
  tags               = try(var.tags, local.core.tags)
  vpc_id             = local.core.vpc_id
  vpc_cidrs          = local.core.vpc_cidrs
  subnet_ids_by_tier = local.core.subnet_ids_by_tier
  kubernetes_cluster = var.kubernetes_cluster
}

module "karpenter" {
  count = var.features.karpenter ? 1 : 0

  source       = "../../../../modules/aws/karpenter-infra"
  name_prefix  = var.name_prefix
  location     = var.location
  tags         = try(var.tags, local.core.tags)
  cluster_name = module.eks[0].cluster_name
  karpenter    = var.kubernetes_cluster.karpenter

  depends_on = [module.eks]
}

locals {
  pod_identity_entries = try({
    for e in var.kubernetes_cluster.pod_identities : e.name => e
  }, {})
}

module "pod_identity" {
  count = var.features.pod_identities && length(local.pod_identity_entries) > 0 ? 1 : 0

  source       = "../../../../modules/aws/pod-identity"
  cluster_name = module.eks[0].cluster_name
  tags         = try(var.tags, local.core.tags)
  entries = {
    for k, v in local.pod_identity_entries : k => {
      namespace       = v.namespace
      service_account = v.service_account
      policy_arns     = try(v.policy_arns, [])
      policy_json     = try(v.policy_json, "")
    }
  }

  depends_on = [module.eks]
}