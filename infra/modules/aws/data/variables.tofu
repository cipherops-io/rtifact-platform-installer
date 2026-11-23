variable "name_prefix" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "features" {
  type = object({
    postgresql         = bool
    redis              = bool
    object_storage     = bool
    container_registry = bool
  })
}

variable "vpc_id" {
  type = string
}

variable "vpc_cidrs" {
  description = "CIDRs allowed to reach PostgreSQL and Redis."
  type        = list(string)
}

variable "subnet_ids_by_tier" {
  type = map(list(string))
}

variable "kms_key_arn" {
  description = "Customer-managed key for encryption; AWS-managed keys are used when null."
  type        = string
  default     = null
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
