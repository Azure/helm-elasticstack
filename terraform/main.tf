provider "azurerm" {
  subscription_id = "${var.azure_subscription_id}"
  client_id = "${var.azure_client_id}"
  client_secret = "${var.azure_client_secret}"
  tenant_id = "${var.azure_tenant_id}"
}

resource "azurerm_resource_group" "elasticsearch-env" {
  location = "${var.azure_location}"
  name = "pltsrv-${var.cluster_enviroment}-elk-${var.azure_location}-rg"
  
  tags {
    environment = "${var.azure_location}",
    business_unit = "platform-services",
    function = "monitoring"
  }
}

resource "azurerm_virtual_network" "elasticsearch_vnet" {
  name                = "pltsrv-${var.cluster_enviroment}-elk-${var.azure_location}-vnet"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.elasticsearch-env.name}"
  address_space       = ["10.1.0.0/24"]
}

resource "azurerm_subnet" "elasticsearch_subnet" {
  name                 = "pltsrv-${var.cluster_environment}-elk-${var.azure_location}-subnet"
  resource_group_name  = "${azurerm_resource_group.elasticsearch-env.name}"
  virtual_network_name = "${azurerm_virtual_network.elasticsearch_vnet.name}"
  address_prefix       = "10.1.0.0/24"
}

resource "azurerm_key_vault" "elasticsearch_keyvault" {
  name                = "pltsrv-${var.cluster_environment}-elk-${var.azure_location}-keyvault"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.elasticsearch-env.name}"

  sku {
    name = "basic"
  }

  tenant_id = "${var.azure_tenant_id}"

  access_policy {
    tenant_id = "${var.azure_tenant_id}"
    object_id = "${var.user_object_id}"

    key_permissions = [
      "get",
      "list",
      "set",
      "delete",
      "Backup",
    ]

    secret_permissions = [
      "get",
      "list",
      "set",
      "delete",
      "backup",
    ]
  }

  enabled_for_disk_encryption = true

  tags {
    environment = "${var.azure_location}",
    business_unit = "platform-services",
    function = "monitoring"
  }
}