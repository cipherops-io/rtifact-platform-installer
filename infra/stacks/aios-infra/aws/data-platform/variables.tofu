variable "account_id" {
  type = string
}

variable "location" {
  type = string
}

variable "name_prefix" {
  type = string
}

variable "workload_environment" {
  type    = string
  default = null
}

variable "features" { type = any }

variable "tags" {
  type    = map(string)
  default = {}
}

variable "upstream_state" {
  type    = any
  default = {}
}

variable "kms" {
  type = object({
    deletion_window_days = optional(number, 30)
  })
  default = {}
}

variable "postgresql" {
  type = object({
    engine_version           = optional(string, "16")
    instance_class           = optional(string, "db.t4g.medium")
    allocated_storage_gb     = optional(number, 100)
    max_allocated_storage_gb = optional(number, 500)
    multi_az                 = optional(bool, false)
    backup_retention_days    = optional(number, 7)
    deletion_protection      = optional(bool, true)
    subnet_tier              = optional(string, "datastore")
  })
  default = {}
}

variable "redis" {
  type = object({
    engine_version     = optional(string, "7.1")
    node_type          = optional(string, "cache.t4g.small")
    num_cache_clusters = optional(number, 1)
    subnet_tier        = optional(string, "datastore")
  })
  default = {}
}

variable "object_storage" {
  type = object({
    versioning = optional(bool, true)
  })
  default = {}
}

variable "container_registry" {
  type = object({
    image_tag_mutability = optional(string, "IMMUTABLE")
    scan_on_push         = optional(bool, true)
  })
  default = {}
}
