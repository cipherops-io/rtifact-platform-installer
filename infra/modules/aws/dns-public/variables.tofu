variable "name_prefix" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}
variable "public_dns" { type = any }

variable "delegation_ttl" {
  description = "TTL of the NS delegation record in the parent zone (public_dns.delegation_ttl)."
  type        = number
  default     = 300
  nullable    = false

  validation {
    condition     = var.delegation_ttl >= 0 && var.delegation_ttl <= 2147483647
    error_message = "public_dns.delegation_ttl must be between 0 and 2147483647 seconds."
  }
}
