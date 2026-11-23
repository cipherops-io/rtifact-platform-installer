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

variable "vnet_cidr" {
  type = string

  validation {
    condition     = can(cidrhost(var.vnet_cidr, 0))
    error_message = "network.vnet_cidr must be an IPv4 CIDR block."
  }
}

variable "node_subnet_cidr" {
  type = string

  validation {
    condition     = can(cidrhost(var.node_subnet_cidr, 0))
    error_message = "network.node_subnet_cidr must be an IPv4 CIDR block."
  }
}

variable "load_balancer_subnet_cidr" {
  type = string

  validation {
    condition     = can(cidrhost(var.load_balancer_subnet_cidr, 0))
    error_message = "network.load_balancer_subnet_cidr must be an IPv4 CIDR block."
  }
}

variable "private_endpoint_subnet_cidr" {
  type = string

  validation {
    condition     = can(cidrhost(var.private_endpoint_subnet_cidr, 0))
    error_message = "network.private_endpoint_subnet_cidr must be an IPv4 CIDR block."
  }
}

variable "management_subnet_cidr" {
  type = string

  validation {
    condition     = can(cidrhost(var.management_subnet_cidr, 0))
    error_message = "network.management_subnet_cidr must be an IPv4 CIDR block."
  }
}

variable "features" {
  type = object({
    azure_firewall              = optional(bool, false)
    azure_vpn_gateway           = optional(bool, false)
    azure_express_route_gateway = optional(bool, false)
    private_dns_resolver        = optional(bool, false)
  })
  default = {
    azure_firewall              = false
    azure_vpn_gateway           = false
    azure_express_route_gateway = false
    private_dns_resolver        = false
  }
}

variable "azure_firewall_subnet_cidr" {
  type    = string
  default = null

  validation {
    condition     = var.azure_firewall_subnet_cidr == null || can(cidrhost(var.azure_firewall_subnet_cidr, 0))
    error_message = "network.azure_firewall_subnet_cidr must be an IPv4 CIDR block."
  }
}

variable "gateway_subnet_cidr" {
  type    = string
  default = null

  validation {
    condition     = var.gateway_subnet_cidr == null || can(cidrhost(var.gateway_subnet_cidr, 0))
    error_message = "network.gateway_subnet_cidr must be an IPv4 CIDR block."
  }
}

variable "azure_bastion_subnet_cidr" {
  type    = string
  default = null

  validation {
    condition     = var.azure_bastion_subnet_cidr == null || can(cidrhost(var.azure_bastion_subnet_cidr, 0))
    error_message = "network.azure_bastion_subnet_cidr must be an IPv4 CIDR block."
  }
}

variable "dns_resolver_inbound_subnet_cidr" {
  type    = string
  default = null

  validation {
    condition     = var.dns_resolver_inbound_subnet_cidr == null || can(cidrhost(var.dns_resolver_inbound_subnet_cidr, 0))
    error_message = "network.dns_resolver_inbound_subnet_cidr must be an IPv4 CIDR block."
  }
}

variable "dns_resolver_outbound_subnet_cidr" {
  type    = string
  default = null

  validation {
    condition     = var.dns_resolver_outbound_subnet_cidr == null || can(cidrhost(var.dns_resolver_outbound_subnet_cidr, 0))
    error_message = "network.dns_resolver_outbound_subnet_cidr must be an IPv4 CIDR block."
  }
}

variable "bastion_managed" {
  type    = bool
  default = false
}

variable "nat" {
  description = "NAT gateway for the node subnet and its public IP (network.nat in customer YAML)."
  type = object({
    sku_name = optional(string, "Standard")
    public_ip = optional(object({
      sku               = optional(string, "Standard")
      allocation_method = optional(string, "Static")
    }), {})
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Standard", "StandardV2"], var.nat.sku_name)
    error_message = "network.nat.sku_name must be Standard or StandardV2."
  }
  validation {
    condition     = var.nat.public_ip.sku == var.nat.sku_name
    error_message = "network.nat.public_ip.sku must match network.nat.sku_name (a StandardV2 NAT gateway needs StandardV2 public IPs)."
  }
  validation {
    condition     = var.nat.public_ip.allocation_method == "Static"
    error_message = "network.nat.public_ip.allocation_method must be Static for Standard and StandardV2 public IPs."
  }
}

variable "dns_resolver_inbound" {
  description = "Inbound endpoint IP configuration for the private DNS resolver (private_dns_resolver in customer YAML)."
  type = object({
    enabled                      = optional(bool, true)
    private_ip_allocation_method = optional(string, "Dynamic")
    private_ip_address           = optional(string)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Dynamic", "Static"], var.dns_resolver_inbound.private_ip_allocation_method)
    error_message = "private_dns_resolver.inbound_private_ip_allocation_method must be Dynamic or Static."
  }
  validation {
    condition     = var.dns_resolver_inbound.private_ip_allocation_method == "Dynamic" || can(cidrhost("${coalesce(var.dns_resolver_inbound.private_ip_address, "invalid")}/32", 0))
    error_message = "private_dns_resolver.inbound_private_ip_address must be an IPv4 address when the allocation method is Static."
  }
}
variable "dns_resolver_outbound_enabled" {
  description = "Create the private DNS resolver outbound endpoint (private_dns_resolver.outbound_endpoint_enabled)."
  type        = bool
  default     = true
  nullable    = false
}
