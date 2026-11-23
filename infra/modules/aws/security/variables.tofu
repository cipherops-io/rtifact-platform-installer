variable "name_prefix" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "deletion_window_days" {
  type     = number
  default  = 30
  nullable = false

  validation {
    condition     = var.deletion_window_days >= 7 && var.deletion_window_days <= 30
    error_message = "kms.deletion_window_days must be between 7 and 30."
  }
}

variable "enable_key_rotation" {
  description = "Rotate the data services KMS key automatically (kms.enable_key_rotation)."
  type        = bool
  default     = true
  nullable    = false
}

variable "rotation_period_in_days" {
  description = "Days between automatic rotations (kms.rotation_period_in_days); AWS default is 365."
  type        = number
  default     = null

  validation {
    condition     = var.rotation_period_in_days == null || (var.rotation_period_in_days >= 90 && var.rotation_period_in_days <= 2560)
    error_message = "kms.rotation_period_in_days must be between 90 and 2560."
  }
  validation {
    condition     = var.rotation_period_in_days == null || var.enable_key_rotation
    error_message = "kms.rotation_period_in_days requires kms.enable_key_rotation = true."
  }
}
