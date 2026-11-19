variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "name_prefix" { type = string }
variable "subnet_id" { type = string }
variable "sku_name" {
  type    = string
  default = "WAF_v2"
}
variable "sku_tier" {
  type    = string
  default = "WAF_v2"
}
variable "capacity" {
  type    = number
  default = 2
}
variable "availability_zones" {
  type    = list(string)
  default = []
}
variable "tags" {
  type    = map(string)
  default = {}
}