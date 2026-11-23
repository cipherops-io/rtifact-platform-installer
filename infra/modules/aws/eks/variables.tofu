variable "name_prefix" { type = string }
variable "location" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "vpc_id" { type = string }
variable "vpc_cidrs" { type = list(string) }
variable "subnet_ids_by_tier" { type = map(list(string)) }

variable "kubernetes_cluster" {
  description = "kubernetes_cluster block from the customer config; optional keys fall back to the defaults in main.tofu."
  type        = any

  validation {
    condition     = contains(["API", "API_AND_CONFIG_MAP"], try(var.kubernetes_cluster.authentication_mode, "API_AND_CONFIG_MAP"))
    error_message = "kubernetes_cluster.authentication_mode must be API or API_AND_CONFIG_MAP (access entries need the API mode)."
  }
  validation {
    condition     = contains(["", "STANDARD", "EXTENDED"], try(var.kubernetes_cluster.support_type, ""))
    error_message = "kubernetes_cluster.support_type must be STANDARD or EXTENDED."
  }
  validation {
    condition     = can(tobool(try(var.kubernetes_cluster.bootstrap_cluster_creator_admin_permissions, false)))
    error_message = "kubernetes_cluster.bootstrap_cluster_creator_admin_permissions must be true or false."
  }
  validation {
    condition     = contains(["AmazonEKSClusterAdminPolicy", "AmazonEKSAdminPolicy", "AmazonEKSEditPolicy", "AmazonEKSViewPolicy"], try(var.kubernetes_cluster.admin_access.policy, "AmazonEKSClusterAdminPolicy"))
    error_message = "kubernetes_cluster.admin_access.policy must be AmazonEKSClusterAdminPolicy, AmazonEKSAdminPolicy, AmazonEKSEditPolicy or AmazonEKSViewPolicy."
  }
  validation {
    condition = (
      try(var.kubernetes_cluster.admin_access.scope, "cluster") == "cluster" ||
      (try(var.kubernetes_cluster.admin_access.scope, "") == "namespace" && length(try(var.kubernetes_cluster.admin_access.namespaces, [])) > 0)
    )
    error_message = "kubernetes_cluster.admin_access.scope must be cluster, or namespace with a non-empty admin_access.namespaces list."
  }
  validation {
    condition = alltrue([
      for a in values(try(var.kubernetes_cluster.addons, {})) :
      contains(["NONE", "OVERWRITE"], try(a.resolve_conflicts_on_create, "OVERWRITE")) &&
      contains(["NONE", "OVERWRITE", "PRESERVE"], try(a.resolve_conflicts_on_update, "OVERWRITE"))
    ])
    error_message = "kubernetes_cluster.addons.*.resolve_conflicts_on_create must be NONE or OVERWRITE; resolve_conflicts_on_update must be NONE, OVERWRITE or PRESERVE."
  }
  validation {
    condition = alltrue([
      for g in values(try(var.kubernetes_cluster.node_groups, {})) :
      contains(["ON_DEMAND", "SPOT"], g.capacity_type) &&
      contains(["AL2_x86_64", "AL2_x86_64_GPU", "AL2_ARM_64", "CUSTOM", "BOTTLEROCKET_ARM_64", "BOTTLEROCKET_x86_64", "BOTTLEROCKET_ARM_64_FIPS", "BOTTLEROCKET_x86_64_FIPS", "BOTTLEROCKET_ARM_64_NVIDIA", "BOTTLEROCKET_x86_64_NVIDIA", "BOTTLEROCKET_ARM_64_NVIDIA_FIPS", "BOTTLEROCKET_x86_64_NVIDIA_FIPS", "WINDOWS_CORE_2019_x86_64", "WINDOWS_FULL_2019_x86_64", "WINDOWS_CORE_2022_x86_64", "WINDOWS_FULL_2022_x86_64", "WINDOWS_CORE_2025_x86_64", "WINDOWS_FULL_2025_x86_64", "AL2023_x86_64_STANDARD", "AL2023_ARM_64_STANDARD", "AL2023_x86_64_NEURON", "AL2023_x86_64_NVIDIA", "AL2023_ARM_64_NVIDIA"], try(g.ami_type, "AL2023_x86_64_STANDARD")) &&
      can(regex("^/dev/[a-z0-9]+$", try(g.disk.device_name, "/dev/xvda"))) &&
      alltrue([for t in try(g.taints, []) : contains(["NO_SCHEDULE", "NO_EXECUTE", "PREFER_NO_SCHEDULE"], t.effect) && length(t.key) <= 63 && length(try(t.value, "")) <= 63]) &&
      contains(["standard", "gp2", "gp3", "io1", "io2", "sc1", "st1"], g.disk.type) &&
      contains(["enabled", "disabled"], try(g.metadata.http_endpoint, "enabled")) &&
      contains(["optional", "required"], g.metadata.http_tokens) &&
      g.metadata.hop_limit >= 1 && g.metadata.hop_limit <= 64 &&
      g.min_size >= 0 && g.min_size <= g.desired_size && g.desired_size <= g.max_size && g.max_size >= 1
    ])
    error_message = "kubernetes_cluster.node_groups: capacity_type must be ON_DEMAND or SPOT; ami_type an EKS AMI type (for example AL2023_x86_64_STANDARD); disk.device_name a /dev path; taints need key/value up to 63 chars and effect NO_SCHEDULE, NO_EXECUTE or PREFER_NO_SCHEDULE; disk.type a valid EBS type; metadata.http_endpoint enabled/disabled; metadata.http_tokens optional/required; metadata.hop_limit 1-64; and 0 <= min_size <= desired_size <= max_size (max_size >= 1)."
  }
}
