variable "name_prefix" { type = string }
variable "location" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "vpc_id" { type = string }

variable "subnet_id" { type = string }

variable "bastion" {
  type = object({
    access          = optional(string, "ssm")
    instance_type   = optional(string, "t3.medium")
    public_ip       = optional(bool, false)
    allowed_cidrs   = optional(list(string), [])
    ssh_port        = optional(number, 22)
    ssh_public_key  = optional(string, "")
    admin_username  = optional(string, "ec2-user")
    architecture    = optional(string, "x86_64")
    ami_id          = optional(string, "")
    iam_policy_arns = optional(list(string), ["arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"])
    ssh_key = optional(object({
      algorithm = optional(string, "RSA")
      rsa_bits  = optional(number, 4096)
    }), {})
    metadata = optional(object({
      http_endpoint = optional(string, "enabled")
      http_tokens   = optional(string, "required")
      hop_limit     = optional(number, 2)
    }), {})
    disk = optional(object({
      size_gb               = optional(number, 100)
      type                  = optional(string, "gp3")
      encrypted             = optional(bool, true)
      delete_on_termination = optional(bool, true)
    }), {})
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["ssm", "ssh"], var.bastion.access)
    error_message = "bastion.access must be ssm or ssh."
  }
  validation {
    condition     = can(regex("^[a-z][a-z0-9-]*[0-9][a-z0-9-]*\\.[a-z0-9]+$", var.bastion.instance_type))
    error_message = "bastion.instance_type must be an EC2 instance type such as t3.medium."
  }
  validation {
    condition     = alltrue([for c in var.bastion.allowed_cidrs : can(cidrhost(c, 0)) && c != "0.0.0.0/0"])
    error_message = "bastion.allowed_cidrs must be valid CIDRs and must not include 0.0.0.0/0."
  }
  validation {
    condition     = var.bastion.access != "ssh" || length(var.bastion.allowed_cidrs) > 0
    error_message = "bastion.access ssh requires at least one bastion.allowed_cidrs entry."
  }
  validation {
    condition     = var.bastion.ssh_port >= 1 && var.bastion.ssh_port <= 65535 && floor(var.bastion.ssh_port) == var.bastion.ssh_port
    error_message = "bastion.ssh_port must be a whole number between 1 and 65535."
  }
  validation {
    condition     = var.bastion.ssh_public_key == "" || can(regex("^(ssh-rsa|ssh-ed25519|ecdsa-sha2-nistp(256|384|521)) [A-Za-z0-9+/=]+( .*)?$", trimspace(var.bastion.ssh_public_key)))
    error_message = "bastion.ssh_public_key must be an OpenSSH public key (ssh-rsa, ssh-ed25519 or ecdsa-sha2-*)."
  }
  validation {
    condition     = can(regex("^[a-z_][a-z0-9_-]{0,31}$", var.bastion.admin_username))
    error_message = "bastion.admin_username must be a Linux user name (lowercase, max 32 characters)."
  }
  validation {
    condition     = contains(["x86_64", "arm64"], var.bastion.architecture)
    error_message = "bastion.architecture must be x86_64 or arm64."
  }
  validation {
    condition     = can(regex("^(ami-[0-9a-f]{8,17})?$", var.bastion.ami_id))
    error_message = "bastion.ami_id must be empty (latest Amazon Linux 2023) or an AMI ID such as ami-0123456789abcdef0."
  }
  validation {
    condition     = alltrue([for a in var.bastion.iam_policy_arns : can(regex("^arn:aws[a-z-]*:iam::(aws|[0-9]{12}):policy/.+$", a))])
    error_message = "bastion.iam_policy_arns must be IAM policy ARNs."
  }
  validation {
    condition     = contains(["RSA", "ED25519"], var.bastion.ssh_key.algorithm)
    error_message = "bastion.ssh_key.algorithm must be RSA or ED25519."
  }
  validation {
    condition     = contains([2048, 3072, 4096], var.bastion.ssh_key.rsa_bits)
    error_message = "bastion.ssh_key.rsa_bits must be 2048, 3072 or 4096."
  }
  validation {
    condition     = contains(["enabled", "disabled"], var.bastion.metadata.http_endpoint)
    error_message = "bastion.metadata.http_endpoint must be enabled or disabled."
  }
  validation {
    condition     = contains(["optional", "required"], var.bastion.metadata.http_tokens)
    error_message = "bastion.metadata.http_tokens must be optional or required."
  }
  validation {
    condition     = var.bastion.metadata.hop_limit >= 1 && var.bastion.metadata.hop_limit <= 64
    error_message = "bastion.metadata.hop_limit must be between 1 and 64."
  }
  validation {
    condition     = var.bastion.access != "ssm" || var.bastion.metadata.http_endpoint == "enabled"
    error_message = "bastion.access ssm needs bastion.metadata.http_endpoint = enabled (the SSM agent reads instance credentials from IMDS)."
  }
  validation {
    condition     = var.bastion.disk.size_gb >= 8 && var.bastion.disk.size_gb <= 16384
    error_message = "bastion.disk.size_gb must be between 8 and 16384."
  }
  validation {
    condition     = contains(["standard", "gp2", "gp3", "io1", "io2", "sc1", "st1"], var.bastion.disk.type)
    error_message = "bastion.disk.type must be standard, gp2, gp3, io1, io2, sc1 or st1."
  }
}
