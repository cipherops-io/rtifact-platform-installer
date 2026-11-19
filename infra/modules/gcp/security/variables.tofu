variable "name_prefix" { type = string }
variable "location" { type = string }
variable "scope_name" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "sku" {
  type    = string
  default = "standard"
}
