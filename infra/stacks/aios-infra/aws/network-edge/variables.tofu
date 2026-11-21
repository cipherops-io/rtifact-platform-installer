variable "account_id" {
  type = string
}

variable "location" {
  type = string
}

variable "name_prefix" {
  type = string
}

variable "workload_environment" {
  type    = string
  default = null
}

variable "features" { type = any }

variable "tags" {
  type    = map(string)
  default = {}
}

variable "public_dns" {
  type    = any
  default = {}
}

variable "certificates" {
  type    = any
  default = {}
}

variable "upstream_state" {
  type    = any
  default = {}
}

variable "dns_delegation_verified" {
  type    = bool
  default = false
}