variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "name_prefix" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}
variable "firewall_subnet_id" { type = string }
variable "sku_tier" {
  type    = string
  default = "Standard"
}