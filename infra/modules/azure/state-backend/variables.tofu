variable "name_prefix" { type = string }
variable "location" { type = string }
variable "subscription_id" { type = string }
variable "resource_group_name" { type = string }

variable "create_resource_group" {
  type        = bool
  description = "When false, use an existing resource group (e.g. created by core-network or a prior bootstrap)."
  default     = true
}