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
    resource_group              = bool
    virtual_network             = bool
    key_vault                   = bool
    key_vault_private_endpoint  = bool
    postgresql_flexible_server  = bool
    container_registry          = bool
    storage_account             = bool
    redis_cache                 = bool
    azure_managed_redis         = optional(bool, false)
    kubernetes_cluster          = bool
    azure_bastion               = bool
    private_dns_zones           = bool
    azure_files_share           = optional(bool, false)
    communication_email         = optional(bool, false)
    azure_firewall              = optional(bool, false)
    azure_vpn_gateway           = optional(bool, false)
    azure_express_route_gateway = optional(bool, false)
    private_dns_resolver        = optional(bool, false)
    public_dns_zone             = optional(bool, false)
    application_gateway         = optional(bool, false)
    disk_encryption_set         = optional(bool, false)
    ai_foundry                  = optional(bool, false)
  })
}

variable "network" {
  type = object({
    vnet_cidr                         = string
    node_subnet_cidr                  = string
    load_balancer_subnet_cidr         = string
    private_endpoint_subnet_cidr      = string
    management_subnet_cidr            = string
    pod_cidr                          = string
    service_cidr                      = string
    dns_service_ip                    = string
    azure_firewall_subnet_cidr        = optional(string)
    gateway_subnet_cidr               = optional(string)
    azure_bastion_subnet_cidr         = optional(string)
    dns_resolver_inbound_subnet_cidr  = optional(string)
    dns_resolver_outbound_subnet_cidr = optional(string)
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
    sku_name              = string
    version               = string
    storage_mb            = number
    iops                  = optional(number)
    backup_retention_days = optional(number, 7)
    high_availability = optional(object({
      enabled = bool
      mode    = optional(string, "ZoneRedundant")
    }), { enabled = false, mode = "ZoneRedundant" })
    read_replica = optional(object({
      enabled = bool
    }), { enabled = false })
  })
}

variable "redis_cache" {
  type = object({
    sku_name            = optional(string, "Premium")
    family              = optional(string, "P")
    capacity            = optional(number, 1)
    minimum_tls_version = optional(string, "1.2")
  })
  default = {
    sku_name            = "Premium"
    family              = "P"
    capacity            = 1
    minimum_tls_version = "1.2"
  }
}

variable "azure_files" {
  type = object({
    share_name     = optional(string, "aios-shared")
    quota_gb       = optional(number, 1024)
    zone_redundant = optional(bool, true)
  })
  default = {
    share_name     = "aios-shared"
    quota_gb       = 1024
    zone_redundant = true
  }
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
# Fragment merged into each Azure stack variables.tofu (reference only).

variable "bastion" {
  type = object({
    managed         = bool
    vm_size         = optional(string, "Standard_D4s_v5")
    os_disk_size_gb = optional(number, 100)
    os_disk_type    = optional(string, "Premium_LRS")
  })
  default = {
    managed         = false
    vm_size         = "Standard_D4s_v5"
    os_disk_size_gb = 100
    os_disk_type    = "Premium_LRS"
  }
}

variable "azure_firewall" {
  type = object({
    sku_tier = optional(string, "Standard")
  })
  default = { sku_tier = "Standard" }
}

variable "azure_vpn" {
  type = object({
    sku           = optional(string, "VpnGw2")
    gateway_type  = optional(string, "Vpn")
    vpn_type      = optional(string, "RouteBased")
    active_active = optional(bool, false)
  })
  default = {
    sku           = "VpnGw2"
    gateway_type  = "Vpn"
    vpn_type      = "RouteBased"
    active_active = false
  }
}

variable "azure_express_route" {
  type = object({
    sku                      = optional(string, "ErGw2AZ")
    express_route_circuit_id = optional(string)
  })
  default = { sku = "ErGw2AZ" }
}

variable "private_dns_resolver" {
  type = object({
    inbound_endpoint_enabled  = optional(bool, true)
    outbound_endpoint_enabled = optional(bool, true)
  })
  default = {
    inbound_endpoint_enabled  = true
    outbound_endpoint_enabled = true
  }
}

variable "public_dns" {
  type = object({
    zones = optional(list(object({
      name = string
      records = optional(list(object({
        name   = string
        type   = string
        ttl    = optional(number, 300)
        record = string
      })), [])
    })), [])
  })
  default = { zones = [] }
}

variable "application_gateway" {
  type = object({
    sku_name = optional(string, "WAF_v2")
    sku_tier = optional(string, "WAF_v2")
    capacity = optional(number, 2)
  })
  default = {
    sku_name = "WAF_v2"
    sku_tier = "WAF_v2"
    capacity = 2
  }
}

variable "disk_encryption_set" {
  type = object({
    rotation_to_latest_version = optional(bool, true)
  })
  default = { rotation_to_latest_version = true }
}

variable "azure_managed_redis" {
  type = object({
    sku_name = optional(string, "Balanced_B3")
  })
  default = { sku_name = "Balanced_B3" }
}

variable "ai_foundry" {
  type = object({
    public_network_access = optional(string, "Disabled")
    friendly_name         = optional(string)
  })
  default = {
    public_network_access = "Disabled"
  }
}
variable "upstream_state" {
  description = "terraform_remote_state backend settings for upstream layers, keyed by layer name (generated by the installer)."
  type = map(object({
    backend = string
    config  = map(string)
  }))
}
