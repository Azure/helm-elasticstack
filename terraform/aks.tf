
resource "azurerm_kubernetes_cluster" "aks-cluster" {
  name                = "pltsrv-${var.cluster_environment}-k8s-${var.azure_location}-k8s"
  location            = "${var.azure_location}"
  resource_group_name = "${azurerm_resource_group.elasticsearch-env.name}"
  dns_prefix  = "k8s-${var.cluster_environment}"
  kubernetes_version = "${var.cluster_version}"

  linux_profile {
    admin_username = "${var.agent_username}"
    
    ssh_key {
      key_data = "ssh-rsa ..."
    }
  }

  agent_pool_profile {
    name            = "default"
    count           = 3
    vm_size         = "Standard_D1_v2"
    os_type         = "Linux"
    os_disk_size_gb = 30
  }

  service_principal {
    client_id     = "${var.azure_client_id}"
    client_secret = "${var.azure_client_secret}"
  }

   tags {
    environment = "${var.azure_location}",
    business_unit = "platform-services",
    function = "kubernetes-aks"
  }
}


