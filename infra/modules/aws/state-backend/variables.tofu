variable "name_prefix" { type = string }
variable "location" { type = string }
variable "account_id" { type = string }

variable "state" {
  type = object({
    bucket_name = optional(string)
    encryption  = optional(string, "AES256")
    versioning  = optional(bool, true)
    lock_table = optional(object({
      billing_mode   = optional(string, "PAY_PER_REQUEST")
      read_capacity  = optional(number, 5)
      write_capacity = optional(number, 5)
    }), {})
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["AES256", "aws:kms", "aws:kms:dsse"], var.state.encryption)
    error_message = "state.encryption must be AES256, aws:kms or aws:kms:dsse."
  }
  validation {
    condition     = contains(["PAY_PER_REQUEST", "PROVISIONED"], var.state.lock_table.billing_mode)
    error_message = "state.lock_table.billing_mode must be PAY_PER_REQUEST or PROVISIONED."
  }
  validation {
    condition     = var.state.lock_table.read_capacity >= 1 && var.state.lock_table.write_capacity >= 1
    error_message = "state.lock_table.read_capacity and write_capacity must be at least 1."
  }
}