variable "name_prefix" { type = string }
variable "location" { type = string }
variable "account_id" { type = string }

variable "state" {
  type = object({
    bucket_name = optional(string)
    encryption  = optional(string, "AES256")
    versioning  = optional(bool, true)
  })
  default = {}
}