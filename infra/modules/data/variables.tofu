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
  default = "16"
}

variable "postgresql_sku_name" {
  type    = string
  default = "GP_Standard_D8ds_v5"
}

variable "postgresql_storage_mb" {
  type    = number
  default = 1048576
}

variable "redis_sku_name" {
  type    = string
  default = "Balanced_B5"
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
  default = "Premium"
}

variable "private_endpoint_subnet_id" {
  type = string
}

variable "vnet_id" {
  type = string
}

variable "dns_resource_group_name" {
  type = string
}