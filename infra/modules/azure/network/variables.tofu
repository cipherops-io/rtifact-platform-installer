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
}

variable "node_subnet_cidr" {
  type = string
}

variable "load_balancer_subnet_cidr" {
  type = string
}

variable "private_endpoint_subnet_cidr" {
  type = string
}

variable "management_subnet_cidr" {
  type = string
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
}

variable "gateway_subnet_cidr" {
  type    = string
  default = null
}

variable "azure_bastion_subnet_cidr" {
  type    = string
  default = null
}

variable "dns_resolver_inbound_subnet_cidr" {
  type    = string
  default = null
}

variable "dns_resolver_outbound_subnet_cidr" {
  type    = string
  default = null
}

variable "bastion_managed" {
  type    = bool
  default = false
}