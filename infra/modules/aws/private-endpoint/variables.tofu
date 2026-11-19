variable "name" { type = string }
variable "location" { type = string }
variable "scope_name" { type = string }
variable "subnet_id" { type = string }
variable "vpc_id" { type = string }
variable "service_name" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}
