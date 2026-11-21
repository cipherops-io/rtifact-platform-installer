variable "name_prefix" { type = string }
variable "location" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "cluster_name" {
  type        = string
  description = "EKS cluster name for kubernetes.io/cluster/* subnet tags"
}

variable "network" {
  type = object({
    vpc_cidr           = string
    secondary_cidrs    = optional(list(string), [])
    availability_zones = list(string)
    nat_gateway        = optional(string, "single")
    gateway_endpoints  = optional(list(string), [])
    flow_logs          = optional(bool, false)
    subnets = map(object({
      cidrs               = list(string)
      route               = string
      public_ip_on_launch = optional(bool, false)
      roles               = optional(list(string), [])
    }))
  })
}