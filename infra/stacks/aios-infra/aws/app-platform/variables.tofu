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

variable "network" {
  type    = any
  default = {}
}

variable "kubernetes_cluster" {
  type = any
}

variable "bastion" {
  type    = any
  default = {}
}

variable "upstream_state" {
  type    = any
  default = {}
}