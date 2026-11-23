variable "name" { type = string }
variable "location" { type = string }
variable "scope_name" { type = string }
variable "subnet_id" { type = string }
variable "vpc_id" { type = string }
variable "service_name" { type = string }
variable "tags" {
  type    = map(string)
  default = {}
}

variable "private_dns_enabled" {
  description = "Associate the AWS-managed private hosted zone with the VPC so service hostnames resolve to the endpoint."
  type        = bool
  default     = true
  nullable    = false
}
