resource "azurerm_public_ip" "bastion" {
  name                = "${var.name_prefix}-bastion-pip"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = var.tags
}

resource "azurerm_bastion_host" "main" {
  name                = "${var.name_prefix}-bastion"
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = var.settings.sku
  scale_units         = var.settings.scale_units
  tunneling_enabled   = var.settings.tunneling_enabled
  copy_paste_enabled  = var.settings.copy_paste_enabled
  file_copy_enabled   = var.settings.file_copy_enabled
  ip_connect_enabled  = var.settings.ip_connect_enabled
  tags                = var.tags

  ip_configuration {
    name                 = "primary"
    subnet_id            = var.bastion_subnet_id
    public_ip_address_id = azurerm_public_ip.bastion.id
  }
}
