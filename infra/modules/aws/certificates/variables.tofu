variable "name_prefix" { type = string }
variable "location" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}
variable "certificates" { type = any }
variable "zone_id" { type = string }

variable "validation_record_ttl" {
  description = "TTL of the ACM DNS validation records (certificates.validation_record_ttl)."
  type        = number
  default     = 60
  nullable    = false

  validation {
    condition     = var.validation_record_ttl >= 0 && var.validation_record_ttl <= 172800
    error_message = "certificates.validation_record_ttl must be between 0 and 172800 seconds."
  }
}

variable "key_algorithm" {
  description = "ACM key algorithm (certificates.key_algorithm)."
  type        = string
  default     = "RSA_2048"
  nullable    = false

  validation {
    condition     = contains(["RSA_1024", "RSA_2048", "RSA_3072", "RSA_4096", "EC_prime256v1", "EC_secp384r1", "EC_secp521r1"], var.key_algorithm)
    error_message = "certificates.key_algorithm must be RSA_1024, RSA_2048, RSA_3072, RSA_4096, EC_prime256v1, EC_secp384r1 or EC_secp521r1."
  }
}
