variable "name_prefix" { type = string }
variable "location" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "cluster_name" { type = string }
variable "karpenter" {
  description = "kubernetes_cluster.karpenter block from the customer config; optional keys fall back to the defaults in main.tofu."
  type        = any

  validation {
    condition     = try(var.karpenter.queue_retention_seconds, 300) >= 60 && try(var.karpenter.queue_retention_seconds, 300) <= 1209600
    error_message = "kubernetes_cluster.karpenter.queue_retention_seconds must be between 60 and 1209600."
  }
  validation {
    condition     = can(tobool(try(var.karpenter.queue_managed_sse_enabled, true)))
    error_message = "kubernetes_cluster.karpenter.queue_managed_sse_enabled must be true or false."
  }
  validation {
    condition = alltrue([
      for v in [try(var.karpenter.namespace, "karpenter"), try(var.karpenter.service_account, "karpenter")] :
      can(regex("^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$", v))
    ])
    error_message = "kubernetes_cluster.karpenter.namespace and service_account must be DNS-1123 labels."
  }
}