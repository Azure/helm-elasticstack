output "vault_uri" {
  value = "${data.azurerm_key_vault.elasticsearch_keyvault.vault_uri}"
}