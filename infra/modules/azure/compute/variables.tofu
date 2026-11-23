variable "resource_group_name" {
  type = string
}

variable "location" {
  type = string
}

variable "name_prefix" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "subnet_id" {
  type = string
}

variable "access" {
  description = "How operators reach the bastion VM: private (VNet only), public_ip (public IP plus NSG allow-list) or azure_bastion (Azure Bastion tunnel)."
  type        = string
  default     = "private"
  nullable    = false

  validation {
    condition     = contains(["private", "public_ip", "azure_bastion"], var.access)
    error_message = "bastion.access must be private, public_ip or azure_bastion."
  }
}

variable "allowed_ssh_cidrs" {
  description = "Source CIDRs allowed to SSH to the bastion public IP (bastion.allowed_ssh_cidrs, public_ip access only)."
  type        = list(string)
  default     = []
  nullable    = false

  validation {
    condition     = alltrue([for c in var.allowed_ssh_cidrs : can(cidrhost(c, 0))])
    error_message = "bastion.allowed_ssh_cidrs entries must be IPv4 CIDR blocks (use /32 for a single IP)."
  }
  validation {
    condition     = !contains(var.allowed_ssh_cidrs, "0.0.0.0/0")
    error_message = "bastion.allowed_ssh_cidrs must not contain 0.0.0.0/0; list the operator IPs instead."
  }
  validation {
    condition     = var.access != "public_ip" || length(var.allowed_ssh_cidrs) > 0
    error_message = "bastion.access = public_ip requires at least one bastion.allowed_ssh_cidrs entry."
  }
}

variable "vm_size" {
  type     = string
  default  = "Standard_D4s_v5"
  nullable = false

  validation {
    condition     = can(regex("^(Standard|Basic)_[A-Za-z0-9_-]+$", var.vm_size))
    error_message = "bastion.vm_size must be an Azure VM size such as Standard_D4s_v5."
  }
}

variable "os_disk_size_gb" {
  type     = number
  default  = 100
  nullable = false

  validation {
    condition     = var.os_disk_size_gb >= 30 && var.os_disk_size_gb <= 4095
    error_message = "bastion.os_disk_size_gb must be between 30 and 4095."
  }
}

variable "os_disk_type" {
  type     = string
  default  = "Premium_LRS"
  nullable = false

  validation {
    condition     = contains(["Standard_LRS", "StandardSSD_LRS", "Premium_LRS", "StandardSSD_ZRS", "Premium_ZRS"], var.os_disk_type)
    error_message = "bastion.os_disk_type must be Standard_LRS, StandardSSD_LRS, Premium_LRS, StandardSSD_ZRS or Premium_ZRS."
  }
}

variable "os_disk_caching" {
  type     = string
  default  = "ReadWrite"
  nullable = false

  validation {
    condition     = contains(["None", "ReadOnly", "ReadWrite"], var.os_disk_caching)
    error_message = "bastion.os_disk_caching must be None, ReadOnly or ReadWrite."
  }
}

variable "admin_username" {
  type     = string
  default  = "azureuser"
  nullable = false

  validation {
    condition     = can(regex("^[a-z_][a-z0-9_-]{0,31}$", var.admin_username)) && !contains(["root", "admin", "administrator", "guest", "user", "test"], var.admin_username)
    error_message = "bastion.admin_username must be a lowercase Linux user name (max 32 chars) and not a reserved name such as root or admin."
  }
}

variable "image" {
  description = "Marketplace image for the bastion VM (bastion.image)."
  type = object({
    publisher = optional(string, "Canonical")
    offer     = optional(string, "0001-com-ubuntu-server-jammy")
    sku       = optional(string, "22_04-lts-gen2")
    version   = optional(string, "latest")
  })
  default  = {}
  nullable = false

  validation {
    condition     = alltrue([for v in values(var.image) : length(trimspace(v)) > 0])
    error_message = "bastion.image publisher, offer, sku and version must be non-empty."
  }
}

variable "ssh_key" {
  description = "Generated SSH key for the bastion admin user (bastion.ssh_key). Azure Linux VMs accept RSA (2048+) and ED25519."
  type = object({
    algorithm = optional(string, "RSA")
    rsa_bits  = optional(number, 4096)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["RSA", "ED25519"], var.ssh_key.algorithm)
    error_message = "bastion.ssh_key.algorithm must be RSA or ED25519."
  }
  validation {
    condition     = contains([2048, 3072, 4096], var.ssh_key.rsa_bits)
    error_message = "bastion.ssh_key.rsa_bits must be 2048, 3072 or 4096."
  }
}

variable "private_ip" {
  description = "Private IP of the bastion NIC in the management subnet (bastion.private_ip)."
  type = object({
    allocation_method = optional(string, "Dynamic")
    address           = optional(string)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Dynamic", "Static"], var.private_ip.allocation_method)
    error_message = "bastion.private_ip.allocation_method must be Dynamic or Static."
  }
  validation {
    condition     = var.private_ip.allocation_method == "Dynamic" || can(cidrhost("${coalesce(var.private_ip.address, "invalid")}/32", 0))
    error_message = "bastion.private_ip.address must be an IPv4 address when allocation_method is Static."
  }
}
