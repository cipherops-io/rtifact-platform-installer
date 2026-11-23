resource "azurerm_public_ip" "appgw" {
  name                = "${var.name_prefix}-appgw-pip"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = "Standard"
  zones               = var.availability_zones
  tags                = var.tags
}

resource "azurerm_application_gateway" "main" {
  name                = "${var.name_prefix}-appgw"
  resource_group_name = var.resource_group_name
  location            = var.location
  tags                = var.tags

  sku {
    name     = var.sku_name
    tier     = var.sku_tier
    capacity = var.capacity
  }

  ssl_policy {
    policy_type = var.ssl_policy.policy_type
    policy_name = var.ssl_policy.policy_name
  }

  gateway_ip_configuration {
    name      = "appgw-ip-config"
    subnet_id = var.subnet_id
  }

  frontend_port {
    name = "http"
    port = var.frontend_port
  }

  frontend_ip_configuration {
    name                 = "frontend"
    public_ip_address_id = azurerm_public_ip.appgw.id
  }

  backend_address_pool {
    name = "placeholder"
  }

  backend_http_settings {
    name                  = "placeholder"
    cookie_based_affinity = var.backend_http_settings.cookie_based_affinity
    path                  = var.backend_http_settings.path
    port                  = var.backend_http_settings.port
    protocol              = var.backend_http_settings.protocol
    request_timeout       = var.backend_http_settings.request_timeout
  }

  http_listener {
    name                           = "http"
    frontend_ip_configuration_name = "frontend"
    frontend_port_name             = "http"
    protocol                       = "Http"
  }

  request_routing_rule {
    name                       = "default"
    rule_type                  = "Basic"
    priority                   = var.routing_rule_priority
    http_listener_name         = "http"
    backend_address_pool_name  = "placeholder"
    backend_http_settings_name = "placeholder"
  }
}