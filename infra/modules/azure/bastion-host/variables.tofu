variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "name_prefix" { type = string }
variable "bastion_subnet_id" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "settings" {
  description = "Azure Bastion host settings (bastion.azure_bastion in customer YAML). tunneling_enabled is required for `az network bastion ssh`."
  type = object({
    sku                = optional(string, "Standard")
    scale_units        = optional(number, 2)
    tunneling_enabled  = optional(bool, true)
    copy_paste_enabled = optional(bool, true)
    file_copy_enabled  = optional(bool, false)
    ip_connect_enabled = optional(bool, false)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Basic", "Standard", "Premium"], var.settings.sku)
    error_message = "bastion.azure_bastion.sku must be Basic, Standard or Premium (Developer has no subnet and is not supported)."
  }
  validation {
    condition     = var.settings.scale_units >= 2 && var.settings.scale_units <= 50 && floor(var.settings.scale_units) == var.settings.scale_units
    error_message = "bastion.azure_bastion.scale_units must be a whole number between 2 and 50."
  }
  validation {
    condition     = var.settings.sku != "Basic" || (var.settings.scale_units == 2 && !var.settings.tunneling_enabled && !var.settings.file_copy_enabled && !var.settings.ip_connect_enabled)
    error_message = "The Basic Bastion SKU supports none of tunneling_enabled, file_copy_enabled, ip_connect_enabled or scale_units other than 2; use Standard or Premium."
  }
}
