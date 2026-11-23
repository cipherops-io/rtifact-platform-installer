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
    storage_type             = optional(string, "gp3")
    multi_az                 = optional(bool, false)
    backup_retention_days    = optional(number, 7)
    deletion_protection      = optional(bool, true)
    subnet_tier              = optional(string, "datastore")
    port                     = optional(number, 5432)
    db_name                  = optional(string, "aios")
    username                 = optional(string, "aiosadmin")
    publicly_accessible      = optional(bool, false)
    copy_tags_to_snapshot    = optional(bool, true)
    skip_final_snapshot      = optional(bool, false)
  })
  default  = {}
  nullable = false

  validation {
    condition     = can(regex("^[0-9]+(\\.[0-9]+)?$", var.postgresql.engine_version))
    error_message = "postgresql.engine_version must be a major or major.minor version, for example 16 or 16.4."
  }
  validation {
    condition     = can(regex("^db\\.[a-z0-9]+\\.[a-z0-9]+$", var.postgresql.instance_class))
    error_message = "postgresql.instance_class must be an RDS instance class such as db.t4g.medium."
  }
  validation {
    condition     = var.postgresql.allocated_storage_gb >= 20 && var.postgresql.allocated_storage_gb <= 65536
    error_message = "postgresql.allocated_storage_gb must be between 20 and 65536."
  }
  validation {
    condition     = contains(["standard", "gp2", "gp3", "io1", "io2"], var.postgresql.storage_type)
    error_message = "postgresql.storage_type must be standard, gp2, gp3, io1 or io2."
  }
  validation {
    condition     = var.postgresql.backup_retention_days >= 0 && var.postgresql.backup_retention_days <= 35
    error_message = "postgresql.backup_retention_days must be between 0 and 35."
  }
  validation {
    condition     = var.postgresql.port >= 1150 && var.postgresql.port <= 65535
    error_message = "postgresql.port must be between 1150 and 65535."
  }
  validation {
    condition     = can(regex("^[A-Za-z][A-Za-z0-9_]{0,62}$", var.postgresql.db_name))
    error_message = "postgresql.db_name must start with a letter and contain only letters, digits and underscores (max 63)."
  }
  validation {
    condition     = can(regex("^[A-Za-z][A-Za-z0-9_]{0,62}$", var.postgresql.username)) && !contains(["postgres", "rdsadmin", "rdsrepladmin"], lower(var.postgresql.username))
    error_message = "postgresql.username must start with a letter, contain only letters, digits and underscores (max 63), and not be a reserved RDS name."
  }
}

variable "redis" {
  type = object({
    engine                     = optional(string, "redis")
    engine_version             = optional(string, "7.1")
    node_type                  = optional(string, "cache.t4g.small")
    num_cache_clusters         = optional(number, 1)
    subnet_tier                = optional(string, "datastore")
    port                       = optional(number, 6379)
    at_rest_encryption_enabled = optional(bool, true)
    transit_encryption_enabled = optional(bool, true)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["redis", "valkey"], var.redis.engine)
    error_message = "redis.engine must be redis or valkey."
  }
  validation {
    condition     = can(regex("^[0-9]+(\\.[0-9]+)?$", var.redis.engine_version))
    error_message = "redis.engine_version must be a major or major.minor version, for example 7.1."
  }
  validation {
    condition     = can(regex("^cache\\.[a-z0-9]+\\.[a-z0-9]+$", var.redis.node_type))
    error_message = "redis.node_type must be an ElastiCache node type such as cache.t4g.small."
  }
  validation {
    condition     = var.redis.num_cache_clusters >= 1 && var.redis.num_cache_clusters <= 6
    error_message = "redis.num_cache_clusters must be between 1 and 6."
  }
  validation {
    condition     = var.redis.port >= 1 && var.redis.port <= 65535
    error_message = "redis.port must be between 1 and 65535."
  }
}

variable "object_storage" {
  type = object({
    versioning          = optional(bool, true)
    object_ownership    = optional(string, "BucketOwnerEnforced")
    block_public_access = optional(bool, true)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["BucketOwnerEnforced", "BucketOwnerPreferred", "ObjectWriter"], var.object_storage.object_ownership)
    error_message = "object_storage.object_ownership must be BucketOwnerEnforced, BucketOwnerPreferred or ObjectWriter."
  }
}

variable "container_registry" {
  type = object({
    image_tag_mutability = optional(string, "IMMUTABLE")
    scan_on_push         = optional(bool, true)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["MUTABLE", "IMMUTABLE", "IMMUTABLE_WITH_EXCLUSION", "MUTABLE_WITH_EXCLUSION"], var.container_registry.image_tag_mutability)
    error_message = "container_registry.image_tag_mutability must be MUTABLE, IMMUTABLE, IMMUTABLE_WITH_EXCLUSION or MUTABLE_WITH_EXCLUSION."
  }
}
