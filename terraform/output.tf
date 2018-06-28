# Azure Keyvault URI
output "vault_uri" {
  value = "${data.azurerm_key_vault.elasticsearch_keyvault.vault_uri}"
}

#AKS output
output "id" {
    value = "${azurerm_kubernetes_cluster.aks-cluster.id}"
}

output "kube_config" {
  value = "${azurerm_kubernetes_cluster.aks-cluster.kube_config_raw}"
}

output "client_key" {
  value = "${azurerm_kubernetes_cluster.aks-cluster.kube_config.0.client_key}"
}

output "client_certificate" {
  value = "${azurerm_kubernetes_cluster.aks-cluster.kube_config.0.client_certificate}"
}

output "cluster_ca_certificate" {
  value = "${azurerm_kubernetes_cluster.aks-cluster.kube_config.0.cluster_ca_certificate}"
}

output "host" {
  value = "${azurerm_kubernetes_cluster.aks-cluster.kube_config.0.host}"
}