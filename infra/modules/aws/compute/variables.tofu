variable "scope_name" { type = string }
variable "location" { type = string }
variable "name_prefix" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "subnet_id" { type = string }
variable "vm_size" { type = string }
variable "os_disk_size_gb" { type = number }