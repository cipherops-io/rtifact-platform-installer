variable "name_prefix" { type = string }
variable "location" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "vpc_id" { type = string }
variable "vpc_cidrs" { type = list(string) }
variable "subnet_ids_by_tier" { type = map(list(string)) }

variable "kubernetes_cluster" { type = any }