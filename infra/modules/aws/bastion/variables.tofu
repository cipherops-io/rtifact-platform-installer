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
    enabled        = bool
    access         = optional(string, "ssm")
    instance_type  = optional(string, "t3.medium")
    public_ip      = optional(bool, false)
    allowed_cidrs  = optional(list(string), [])
    ssh_public_key = optional(string, "")
    architecture   = optional(string, "x86_64")
    ami_id         = optional(string, "")
    iam_policy_arns = optional(list(string), ["arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"])
    disk = optional(object({
      size_gb   = number
      type      = string
      encrypted = bool
    }), { size_gb = 100, type = "gp3", encrypted = true })
  })
}