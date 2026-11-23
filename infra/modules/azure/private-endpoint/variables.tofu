variable "name" {
  type = string
}

variable "location" {
  type = string
}

variable "resource_group_name" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "private_connection_resource_id" {
  type = string
}

variable "subresource_names" {
  type = list(string)
}

variable "private_dns_zone_ids" {
  type = list(string)
}

variable "connection" {
  description = "Private service connection approval settings (network.private_endpoint in customer YAML)."
  type = object({
    is_manual_connection = optional(bool, false)
    request_message      = optional(string)
  })
  default  = {}
  nullable = false

  validation {
    condition     = var.connection.request_message == null || (var.connection.is_manual_connection && length(coalesce(var.connection.request_message, "")) <= 140)
    error_message = "network.private_endpoint.request_message is only valid with is_manual_connection = true and must be at most 140 characters."
  }
}

variable "tags" {
  type    = map(string)
  default = {}
}