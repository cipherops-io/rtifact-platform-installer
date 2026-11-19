variable "scope_name" {
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

variable "features" {
  type = object({
    postgresql_flexible_server = bool
    container_registry         = bool
    storage_account            = bool
    redis_cache                = bool
    private_dns_zones          = bool
  })
}

variable "postgresql_version" {
  type    = string
  default = "16.4"
}

variable "postgresql_sku_name" {
  type    = string
  default = "db.t3.medium"
}

variable "postgresql_storage_mb" {
  type    = number
  default = 1048576
}

variable "redis_sku_name" {
  type    = string
  default = "cache.t3.small"
}

variable "storage_corpus_gb" {
  type    = number
  default = 250
}

variable "storage_backup_gb" {
  type    = number
  default = 250
}

variable "acr_sku" {
  type    = string
  default = "private"
}

variable "private_endpoint_subnet_id" {
  type = string
}

variable "node_subnet_id" {
  type = string
}

variable "vnet_id" {
  type = string
}

variable "dns_scope_name" {
  type = string
}