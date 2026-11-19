variable "resource_group_name" {
  type = string
}

variable "location" {
  type = string
}

variable "name_prefix" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "kubernetes_version" {
  type = string
}

variable "node_subnet_id" {
  type = string
}

variable "pod_cidr" {
  type = string
}

variable "service_cidr" {
  type = string
}

variable "dns_service_ip" {
  type = string
}

variable "node_pools" {
  type = map(object({
    name       = string
    vm_size    = string
    node_count = number
  }))
  description = "User node pools keyed by role."
}

variable "availability_zones" {
  type        = list(string)
  description = "AKS node pool AZs (centralindia supports 1 and 2 only)."
  default     = ["1", "2", "3"]
}

variable "os_disk_type" {
  type        = string
  description = "AKS node OS disk type: Managed or Ephemeral (not Azure storage SKU names)."
  default     = "Managed"
}
