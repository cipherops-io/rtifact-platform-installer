variable "subscription_id" {
  type = string
}

variable "location" {
  type = string
}

variable "tenant_id" {
  type    = string
  default = null
}

variable "azure_environment" {
  type    = string
  default = "public"
}

variable "use_msi" {
  type    = bool
  default = true
}

variable "name_prefix" {
  type = string
}

variable "workload_environment" {
  type    = string
  default = null
}

variable "features" {
  type = object({
    resource_group               = bool
    virtual_network              = bool
    key_vault                    = bool
    key_vault_private_endpoint   = bool
    postgresql_flexible_server   = bool
    container_registry           = bool
    storage_account              = bool
    redis_cache                  = bool
    kubernetes_cluster           = bool
    azure_bastion                = bool
    private_dns_zones            = bool
  })
}

variable "network" {
  type = object({
    vnet_cidr                    = string
    node_subnet_cidr             = string
    load_balancer_subnet_cidr    = string
    private_endpoint_subnet_cidr = string
    management_subnet_cidr       = string
    pod_cidr                     = string
    service_cidr                 = string
    dns_service_ip               = string
  })
}

variable "kubernetes_cluster" {
  type = object({
    kubernetes_version = string
    availability_zones = list(string)
    node_pools = map(object({
      name       = string
      vm_size    = string
      node_count = number
    }))
  })
}

variable "azure_bastion" {
  type = object({
    vm_size         = string
    os_disk_size_gb = number
  })
  default = {
    vm_size         = "Standard_D2s_v3"
    os_disk_size_gb = 100
  }
}

variable "postgresql_flexible_server" {
  type = object({
    sku_name   = string
    version    = string
    storage_mb = number
    iops       = optional(number)
  })
}

variable "redis_cache" {
  type = object({
    sku_name = string
  })
}

variable "storage_account" {
  type = object({
    corpus_gb = number
    backup_gb = number
  })
}

variable "container_registry" {
  type = object({
    sku = string
  })
}

variable "key_vault" {
  type = object({
    sku = string
  })
}