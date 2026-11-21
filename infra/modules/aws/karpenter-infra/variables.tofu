variable "name_prefix" { type = string }
variable "location" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "cluster_name" { type = string }
variable "karpenter" { type = any }