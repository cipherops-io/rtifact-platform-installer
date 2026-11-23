variable "name_prefix" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "deletion_window_days" {
  type    = number
  default = 30
}
