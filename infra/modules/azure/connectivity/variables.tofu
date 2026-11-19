variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "name_prefix" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}
variable "gateway_subnet_id" { type = string }

variable "vpn_enabled" {
  type    = bool
  default = false
}

variable "express_route_enabled" {
  type    = bool
  default = false
}

variable "vpn" {
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

variable "express_route" {
  type = object({
    sku                      = optional(string, "ErGw2AZ")
    express_route_circuit_id = optional(string)
  })
  default = { sku = "ErGw2AZ" }
}