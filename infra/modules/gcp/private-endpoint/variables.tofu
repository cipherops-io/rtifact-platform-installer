variable "name" { type = string }
variable "location" { type = string }
variable "scope_name" { type = string }
variable "subnet_id" { type = string }
variable "network_id" { type = string }
variable "target_service" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}
