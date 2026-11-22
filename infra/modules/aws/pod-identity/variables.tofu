variable "cluster_name" {
  type = string
}

variable "entries" {
  description = "Pod identity associations keyed by entry name."
  type = map(object({
    namespace       = string
    service_account = string
    policy_arns     = optional(list(string), [])
    policy_json     = optional(string, "")
  }))
}

variable "tags" {
  type    = map(string)
  default = {}
}