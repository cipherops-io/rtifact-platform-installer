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