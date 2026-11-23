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
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Basic", "VpnGw1", "VpnGw2", "VpnGw3", "VpnGw4", "VpnGw5", "VpnGw1AZ", "VpnGw2AZ", "VpnGw3AZ", "VpnGw4AZ", "VpnGw5AZ"], var.vpn.sku)
    error_message = "azure_vpn.sku must be one of Basic, VpnGw1-5 or VpnGw1AZ-5AZ."
  }
  validation {
    condition     = var.vpn.gateway_type == "Vpn"
    error_message = "azure_vpn.gateway_type must be Vpn (ExpressRoute gateways are configured under azure_express_route)."
  }
  validation {
    condition     = contains(["RouteBased", "PolicyBased"], var.vpn.vpn_type)
    error_message = "azure_vpn.vpn_type must be RouteBased or PolicyBased."
  }
  validation {
    condition     = var.vpn.vpn_type == "RouteBased" || var.vpn.sku == "Basic"
    error_message = "A PolicyBased VPN gateway only supports the Basic SKU."
  }
}

variable "express_route" {
  type = object({
    sku                      = optional(string, "ErGw2AZ")
    express_route_circuit_id = optional(string)
    routing_weight           = optional(number, 0)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Standard", "HighPerformance", "UltraPerformance", "ErGwScale", "ErGw1AZ", "ErGw2AZ", "ErGw3AZ"], var.express_route.sku)
    error_message = "azure_express_route.sku must be one of Standard, HighPerformance, UltraPerformance, ErGwScale, ErGw1AZ, ErGw2AZ or ErGw3AZ."
  }
  validation {
    condition     = var.express_route.routing_weight >= 0 && var.express_route.routing_weight <= 32000 && floor(var.express_route.routing_weight) == var.express_route.routing_weight
    error_message = "azure_express_route.routing_weight must be a whole number between 0 and 32000."
  }
}