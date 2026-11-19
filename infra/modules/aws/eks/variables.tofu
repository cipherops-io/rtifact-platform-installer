variable "scope_name" {
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

variable "kubernetes_version" {
  type = string
}

variable "node_subnet_id" {
  type = string
}

variable "pod_cidr" {
  type = string
}

variable "service_cidr" {
  type = string
}

variable "dns_service_ip" {
  type = string
}

variable "availability_zones" {
  type = list(string)
}

variable "node_pools" {
  type = map(object({
    name       = string
    vm_size    = string
    node_count = number
  }))
}