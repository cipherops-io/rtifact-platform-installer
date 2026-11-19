variable "resource_group_name" { type = string }
variable "zones" {
  type = list(object({
    name = string
    records = optional(list(object({
      name   = string
      type   = string
      ttl    = optional(number)
      record = string
    })), [])
  }))
  default = []
}
variable "tags" {
  type    = map(string)
  default = {}
}