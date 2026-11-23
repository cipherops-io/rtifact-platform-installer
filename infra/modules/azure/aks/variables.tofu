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

  validation {
    condition     = can(regex("^[0-9]+\\.[0-9]+(\\.[0-9]+)?$", var.kubernetes_version))
    error_message = "kubernetes_cluster.kubernetes_version must look like 1.30 or 1.30.5."
  }
}

variable "node_subnet_id" {
  type = string
}

variable "pod_cidr" {
  type = string

  validation {
    condition     = can(cidrhost(var.pod_cidr, 0))
    error_message = "network.pod_cidr must be an IPv4 CIDR block."
  }
}

variable "service_cidr" {
  type = string

  validation {
    condition     = can(cidrhost(var.service_cidr, 0))
    error_message = "network.service_cidr must be an IPv4 CIDR block."
  }
}

variable "dns_service_ip" {
  type = string

  validation {
    condition     = can(cidrhost("${var.dns_service_ip}/32", 0))
    error_message = "network.dns_service_ip must be an IPv4 address."
  }
}

variable "node_pools" {
  type = map(object({
    name            = optional(string)
    vm_size         = string
    node_count      = number
    os_disk_size_gb = optional(number, 128)
    mode            = optional(string, "User")
  }))
  description = "Node pools keyed by role. The `system` key overrides system_pool; the rest become user pools."
  nullable    = false

  validation {
    condition     = alltrue([for k, v in var.node_pools : can(regex("^[a-z][a-z0-9]{0,11}$", coalesce(v.name, k)))])
    error_message = "kubernetes_cluster.node_pools names must be lowercase alphanumeric, start with a letter and be at most 12 characters (set name when the key does not qualify)."
  }
  validation {
    condition     = alltrue([for v in values(var.node_pools) : v.node_count >= 0 && v.node_count <= 1000 && floor(v.node_count) == v.node_count])
    error_message = "kubernetes_cluster.node_pools node_count must be a whole number between 0 and 1000."
  }
  validation {
    condition     = alltrue([for v in values(var.node_pools) : v.os_disk_size_gb >= 30 && v.os_disk_size_gb <= 2048])
    error_message = "kubernetes_cluster.node_pools os_disk_size_gb must be between 30 and 2048."
  }
  validation {
    condition     = alltrue([for v in values(var.node_pools) : contains(["System", "User"], v.mode)])
    error_message = "kubernetes_cluster.node_pools mode must be System or User."
  }
}

variable "system_pool" {
  description = "Default (system) node pool used when node_pools has no `system` entry (kubernetes_cluster.system_pool)."
  type = object({
    name            = optional(string, "system")
    vm_size         = optional(string, "Standard_D4s_v5")
    node_count      = optional(number, 3)
    os_disk_size_gb = optional(number, 128)
  })
  default  = {}
  nullable = false

  validation {
    condition     = can(regex("^[a-z][a-z0-9]{0,11}$", var.system_pool.name))
    error_message = "kubernetes_cluster.system_pool.name must be lowercase alphanumeric, start with a letter and be at most 12 characters."
  }
  validation {
    condition     = var.system_pool.node_count >= 1 && var.system_pool.node_count <= 1000
    error_message = "kubernetes_cluster.system_pool.node_count must be between 1 and 1000."
  }
  validation {
    condition     = var.system_pool.os_disk_size_gb >= 30 && var.system_pool.os_disk_size_gb <= 2048
    error_message = "kubernetes_cluster.system_pool.os_disk_size_gb must be between 30 and 2048."
  }
}

variable "availability_zones" {
  type        = list(string)
  description = "AKS node pool AZs (centralindia supports 1 and 2 only)."
  default     = ["1", "2", "3"]
  nullable    = false

  validation {
    condition     = alltrue([for z in var.availability_zones : contains(["1", "2", "3"], z)])
    error_message = "kubernetes_cluster.availability_zones entries must be \"1\", \"2\" or \"3\"."
  }
}

variable "os_disk_type" {
  type        = string
  description = "AKS node OS disk type: Managed or Ephemeral (not Azure storage SKU names)."
  default     = "Managed"
  nullable    = false

  validation {
    condition     = contains(["Managed", "Ephemeral"], var.os_disk_type)
    error_message = "kubernetes_cluster.os_disk_type must be Managed or Ephemeral."
  }
}

variable "cluster" {
  description = "AKS control plane settings (kubernetes_cluster in customer YAML)."
  type = object({
    sku_tier                  = optional(string, "Standard")
    oidc_issuer_enabled       = optional(bool, true)
    workload_identity_enabled = optional(bool, true)
    node_provisioning_mode    = optional(string, "Manual")
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Free", "Standard", "Premium"], var.cluster.sku_tier)
    error_message = "kubernetes_cluster.sku_tier must be Free, Standard or Premium."
  }
  validation {
    condition     = contains(["Manual", "Auto"], var.cluster.node_provisioning_mode)
    error_message = "kubernetes_cluster.node_provisioning_mode must be Manual or Auto."
  }
  validation {
    condition     = !var.cluster.workload_identity_enabled || var.cluster.oidc_issuer_enabled
    error_message = "kubernetes_cluster.workload_identity_enabled requires oidc_issuer_enabled = true."
  }
}

variable "api_server" {
  description = "API server exposure (kubernetes_cluster.api_server). Changing private_cluster_enabled or private_dns_zone_id recreates the cluster."
  type = object({
    private_cluster_enabled = optional(bool, true)
    private_dns_zone_id     = optional(string, "System")
    authorized_ip_ranges    = optional(list(string), [])
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["System", "None"], var.api_server.private_dns_zone_id) || can(regex("^/subscriptions/[^/]+/resourceGroups/[^/]+/providers/Microsoft.Network/privateDnsZones/[^/]+$", var.api_server.private_dns_zone_id))
    error_message = "kubernetes_cluster.api_server.private_dns_zone_id must be System, None or a private DNS zone resource ID."
  }
  validation {
    condition     = alltrue([for c in var.api_server.authorized_ip_ranges : can(cidrhost(c, 0))])
    error_message = "kubernetes_cluster.api_server.authorized_ip_ranges entries must be IPv4 CIDR blocks (use /32 for a single IP)."
  }
  validation {
    condition     = length(var.api_server.authorized_ip_ranges) == 0 || !var.api_server.private_cluster_enabled
    error_message = "kubernetes_cluster.api_server.authorized_ip_ranges only applies to public clusters (private_cluster_enabled = false); reach a private cluster through the bastion."
  }
}

variable "network_profile" {
  description = "AKS network profile (kubernetes_cluster.network_profile)."
  type = object({
    network_plugin      = optional(string, "azure")
    network_plugin_mode = optional(string, "overlay")
    outbound_type       = optional(string, "userAssignedNATGateway")
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["azure", "kubenet", "none"], var.network_profile.network_plugin)
    error_message = "kubernetes_cluster.network_profile.network_plugin must be azure, kubenet or none."
  }
  validation {
    condition     = contains(["overlay", "none"], var.network_profile.network_plugin_mode)
    error_message = "kubernetes_cluster.network_profile.network_plugin_mode must be overlay or none."
  }
  validation {
    condition     = var.network_profile.network_plugin_mode != "overlay" || var.network_profile.network_plugin == "azure"
    error_message = "network_plugin_mode overlay requires network_plugin azure."
  }
  validation {
    condition     = contains(["loadBalancer", "userDefinedRouting", "managedNATGateway", "userAssignedNATGateway", "none"], var.network_profile.outbound_type)
    error_message = "kubernetes_cluster.network_profile.outbound_type must be loadBalancer, userDefinedRouting, managedNATGateway, userAssignedNATGateway or none."
  }
}
