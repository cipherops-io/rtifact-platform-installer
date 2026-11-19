variable "prefix" { type = string }
variable "location" { type = string }
variable "project_id" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

