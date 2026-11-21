variable "name_prefix" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}
variable "public_dns" { type = any }