variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "name_prefix" { type = string }
variable "subnet_id" { type = string }

variable "sku_name" {
  type     = string
  default  = "WAF_v2"
  nullable = false

  validation {
    condition     = contains(["Basic", "Standard_v2", "WAF_v2"], var.sku_name)
    error_message = "application_gateway.sku_name must be Basic, Standard_v2 or WAF_v2 (v1 SKUs are retired)."
  }
}

variable "sku_tier" {
  type     = string
  default  = "WAF_v2"
  nullable = false

  validation {
    condition     = var.sku_tier == var.sku_name
    error_message = "application_gateway.sku_tier must equal sku_name (Basic, Standard_v2 or WAF_v2)."
  }
}

variable "capacity" {
  type     = number
  default  = 2
  nullable = false

  validation {
    condition     = var.sku_name == "Basic" ? (var.capacity >= 1 && var.capacity <= 2) : (var.capacity >= 1 && var.capacity <= 125)
    error_message = "application_gateway.capacity must be 1-2 for Basic and 1-125 for v2 SKUs."
  }
}

variable "availability_zones" {
  type     = list(string)
  default  = []
  nullable = false

  validation {
    condition     = alltrue([for z in var.availability_zones : contains(["1", "2", "3"], z)])
    error_message = "application_gateway availability zones must be \"1\", \"2\" or \"3\"."
  }
}

variable "ssl_policy" {
  description = "Predefined TLS policy (application_gateway.ssl_policy). See https://learn.microsoft.com/azure/application-gateway/application-gateway-ssl-policy-overview."
  type = object({
    policy_type = optional(string, "Predefined")
    policy_name = optional(string, "AppGwSslPolicy20170401S")
  })
  default  = {}
  nullable = false

  validation {
    condition     = var.ssl_policy.policy_type == "Predefined"
    error_message = "application_gateway.ssl_policy.policy_type must be Predefined (Custom policies need cipher_suites this module does not manage)."
  }
  validation {
    condition     = contains(["AppGwSslPolicy20150501", "AppGwSslPolicy20170401", "AppGwSslPolicy20170401S", "AppGwSslPolicy20220101", "AppGwSslPolicy20220101S"], var.ssl_policy.policy_name)
    error_message = "application_gateway.ssl_policy.policy_name must be AppGwSslPolicy20150501, AppGwSslPolicy20170401, AppGwSslPolicy20170401S, AppGwSslPolicy20220101 or AppGwSslPolicy20220101S."
  }
}

variable "frontend_port" {
  description = "Listener port of the placeholder HTTP listener (application_gateway.frontend_port)."
  type        = number
  default     = 80
  nullable    = false

  validation {
    condition     = var.frontend_port >= 1 && var.frontend_port <= 65199
    error_message = "application_gateway.frontend_port must be between 1 and 65199."
  }
}

variable "backend_http_settings" {
  description = "Placeholder backend HTTP settings (application_gateway.backend_http_settings); AGIC replaces them for real routes."
  type = object({
    port                  = optional(number, 80)
    protocol              = optional(string, "Http")
    path                  = optional(string, "/")
    cookie_based_affinity = optional(string, "Disabled")
    request_timeout       = optional(number, 30)
  })
  default  = {}
  nullable = false

  validation {
    condition     = var.backend_http_settings.port >= 1 && var.backend_http_settings.port <= 65535
    error_message = "application_gateway.backend_http_settings.port must be between 1 and 65535."
  }
  validation {
    condition     = contains(["Http", "Https"], var.backend_http_settings.protocol)
    error_message = "application_gateway.backend_http_settings.protocol must be Http or Https."
  }
  validation {
    condition     = startswith(var.backend_http_settings.path, "/")
    error_message = "application_gateway.backend_http_settings.path must start with /."
  }
  validation {
    condition     = contains(["Enabled", "Disabled"], var.backend_http_settings.cookie_based_affinity)
    error_message = "application_gateway.backend_http_settings.cookie_based_affinity must be Enabled or Disabled."
  }
  validation {
    condition     = var.backend_http_settings.request_timeout >= 1 && var.backend_http_settings.request_timeout <= 86400
    error_message = "application_gateway.backend_http_settings.request_timeout must be between 1 and 86400 seconds."
  }
}

variable "routing_rule_priority" {
  description = "Priority of the default request routing rule (application_gateway.routing_rule_priority)."
  type        = number
  default     = 1
  nullable    = false

  validation {
    condition     = var.routing_rule_priority >= 1 && var.routing_rule_priority <= 20000
    error_message = "application_gateway.routing_rule_priority must be between 1 and 20000."
  }
}

variable "tags" {
  type    = map(string)
  default = {}
}
