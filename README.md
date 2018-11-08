# Introduction

[![Build Status](https://travis-ci.org/Azure/helm-elasticstack.svg?branch=master)](https://travis-ci.org/Azure/helm-elasticstack)
[![Go Report Card](https://goreportcard.com/badge/github.com/Azure/helm-elasticstack)](https://goreportcard.com/report/github.com/Azure/helm-elasticstack)

These [Helm](https://github.com/kubernetes/helm) charts bootstrap a production ready [Elastic Stack](https://www.elastic.co/products) service on a Kubernetes cluster managed by [Azure Container Service (AKS)](https://docs.microsoft.com/en-us/azure/aks/intro-kubernetes) and other Azure services.

The following features are included:

* Deployment for [Elasticsearch](https://www.elastic.co/products/elasticsearch), [Kibana](https://www.elastic.co/products/kibana) and [Logstash](https://www.elastic.co/products/logstash) services
* Deployment script which retrieves the secrets and certificates from [Azure Key Vault](https://azure.microsoft.com/en-us/services/key-vault/) and injects them into the Helm charts
* TLS termination and load balancing for Kibana using [NGINX Ingress Controller](https://github.com/kubernetes/ingress-nginx)
* [Azure Active Directory](https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-authentication-scenarios) authentication for Kibana
* Integration with [Azure Redis Cache](https://azure.microsoft.com/en-us/services/cache/) acting as middleware for log events between the Log Appenders and Logstash
* TLS connection between Logstash and Redis Cache handled by [stunnel](https://www.stunnel.org/)
* Support for [Multiple Data Pipelines](https://www.elastic.co/blog/logstash-multiple-pipelines) in Logstash allowing multiple Redis Caches as input (e.g one Redis cluster per environment)
* Installation of a [Curator](https://github.com/elastic/curator) cron job that cleans up daily all indexes which are older than 30 days
* Installation of [Elasticsearch Index Templates](https://www.elastic.co/guide/en/elasticsearch/reference/5.6/indices-templates.html) as a pre-deployment step
* Installation of [Elasticsearch Watches](https://www.elastic.co/guide/en/elasticsearch/reference/5.6/watcher-api.html) as a post deployment step. The watches can be used for alerts and notifications over Microsoft Teams/Slack webhook or email
* Installation of [Elasticsearch x-pack license](https://license.elastic.co/download) as a post deployment step

<!-- TOC -->

- [Introduction](#introduction)
  - [Architecture](#architecture)
  - [Azure Resources](#azure-resources)
    - [Azure Key Vault](#azure-key-vault)
    - [Public Static IP and DNS Domain](#public-static-ip-and-dns-domain)
    - [Redis Cache](#redis-cache)
    - [Application for Azure Active Directory](#application-for-azure-active-directory)
    - [Microsoft Teams/Slack incoming Webhook](#microsoft-teamsslack-incoming-webhook)
  - [Customize Logstash Configuration](#customize-logstash-configuration)
    - [Multiple Data Pipelines](#multiple-data-pipelines)
    - [Indexes Clean Up](#indexes-clean-up)
    - [Index Templates](#index-templates)
    - [Index Watches](#index-watches)
    - [Elasticsearch License](#elasticsearch-license)
  - [Installation](#installation)
    - [NGINX Ingress Controller](#nginx-ingress-controller)
    - [Elasticsearch Cluster](#elasticsearch-cluster)
    - [Kibana and Logstash](#kibana-and-logstash)
    - [Rolling Update](#rolling-update)
  - [Contributing](#contributing)

<!-- /TOC -->

## Architecture

![architecture](images/architecture.png?row=true)

## Azure Resources

A few Azure resources need to be provisioned before proceeding with the Helm charts installation.

### Azure Key Vault

All secrets and certificates used by the charts are stored in an Azure Key Vault. The deployment script is able to fetch them and to inject them further into the charts.

You can create a new Key Vault with default permissions:

```console
az keyvault create --name <KEYVAULT_NAME> --resource-group <RESOURCE_GROUP>
```

It is recommended that you use two different principals to operate the Key Vault:

* A _Security Operator_ who has read/write access to secrets, keys and certificates. This principal should be only used for setting up the Key Vault or rotate the secrets.
* A _Deployment Operator_ who is only able to read secrets. This principal should be used to perform the deployment.

You can configure the access policies for these principals as follows:

```console
az keyvault set-policy --upn <SECURITY_OPERATOR_USER_PRINCIPAL> --name <KEYVAULT_NAME> --resource-group <RESOURCE_GROUP> --certificate-permissions create delete get import list update --key-permissions create delete get import list update --secret-permissions get delete list set

az keyvault set-policy --upn <DEPLOYMENT_OPERATOR_USER_PRINCIPAL> --name <KEYVAULT_NAME> --resource-group <RESOURCE_GROP> --secret-permissions get list
```

### Public Static IP and DNS Domain

You can allocate a public static IP in Azure. This IP will be used to expose Kibana to the world.

```console
az network public-ip create -n <IP_NAME> --resource-groug=<RESOURCE_GROUP> --allocation-method=static --dns-name=<IP_NAME>
```

It is recommended to utilize the resource group where the ACS/AKS cluster is deployed.

At this point, you should register your DNS domain name using the public IP returned by Azure and purchase a TLS public certificate. The public certificate should be typically packaged along with the private key in the PKCS12 format in order to import it into Azure Key Vault. You can achieve this with the following command:

```console
export DOMAIN=<YOUR DNS DOMAIN>
openssl pkcs12 -export -in ${DOMAIN}.cer -inkey ${DOMAIN}.key -out ${DOMAIN}.pfx
```

The PKCS12 certificate can be now imported into Key Vault. You should login with your _Security Operator_ principal by executing the `az login` and then execute the import command. During the import, you have to provide a password that it will be used by Key Vault to encrypt the private key.

```console
az keyvault certificate import --name kibana-certificate --vault-name <KEYVAULT_NAME> -f ${DOMAIN}.pfx --password <PASSWORD> --tags domain=${DOMAIN}
```

The private key password must be also stored in a different secret, such that it can be retrieved by the deployment script.

```console
az keyvault secret set --name kibana-certificate-key-password --vault-name <KEYVAULT_NAME> --value <PASSWORD>
```

### Redis Cache

The Azure Redis Cache is used as a middleware between the Log Appenders and Logstash service. This is quite scalable and it also decouples the Log Appenders from Elastic Stack service. You can use any Log Appender which is able to write log events into Redis.

```console
az redis create --name dev-logscache --location <LOCATION> --resrouce-group <RESOURCE_GROUP> --sku Standard --vm-size C1

```

You have to store one of the Redis Keys in Key Vault.

```console
az keyvault secret set --vault-name <KEYVAULT_NAME> --name logstash-dev-redis-key --value=<REDIS_KEY>
```

### Application for Azure Active Directory

An Azure Active Directory application of type _Web app/API_ is required in order to use the AAD as an identity provider for Kibana. The authentication is provided by [oauth2_proxy](https://github.com/bitly/oauth2_proxy) reverse proxy which is deployed in the same POD as Kibana.

```console
az ad app create --display-name Kibana-App --homepage https://${DOMAIN} --reply-urls https://${DOMAIN}/oauth2/callback --identifier-uris https://${DOMAIN}/kibana-app --password <APPLICATION_SECRET>
```

> Note that you have to replace the _${DOMAIN}_ with the public DNS domain you registered.

After the application was created, you should store the Application ID and secret in Key Vault:

```console
az keyvault secret set --name kibana-oauth-client-id --vault-name <KEYVAULT_NAME> --value <APPLICATION_ID>
az keyvault secret set --name kibana-oauth-client-secret --vault-name <KEYVAULT_NAME> --value <APPLICATION_SECRET>
```

In addition to the application credentials, the `oauth2_proxy` reverse proxy requires a random secret which is used to sign the single-sign-on cookie after login. You can generate and store these secret in Key Vault as follows:

```console
cookie_secret=$(openssl rand -hex 64)
az keyvault secret set --name  kibana-oauth-cookie-secret --vault-name <KEYVAULT_NAME> --value ${cookie_secret}
```

You should also update the access list with the emails of the users from your organization which are allowed to access Kibana. The white list is in [oauth2-proxy-config-secret.yaml](charts/kibana-logstash/templates/secrets/oauth2-proxy-config-secret.yaml) file.

### Microsoft Teams/Slack incoming Webhook

The [Elasticsearch Watcher](https://www.elastic.co/guide/en/elasticsearch/reference/master/watcher-api.html) can post notifications into a webhook. For example, you can use a Microsoft Teams webhook, which can be created following these [instructions](https://docs.microsoft.com/en-us/microsoftteams/platform/concepts/connectors). 

After you configured the webhook, you should store its address in Key Vault:

```console
az keyvault secret set --vault-name <KEYVAULT_NAME> -n elasticsearch-watcher-webhook-teams --value "/webhook/<WEBHOOK_URIS>"
```

If you want instead to use a [Slack Incoming Webhook](https://api.slack.com/incoming-webhooks), you can adjust the configuration in the [post-install-watches-secret.yaml](charts/kibana-logstash/templates/post-install-watches-secret.yaml) file.

## Customize Logstash Configuration

### Multiple Data Pipelines

Multiple data pipelines can be defined in the [values.yaml](charts/kibana-logstash/environments/acs/values.yaml) file by creating multiple `stunnel` connections as follows:

```yaml
stunnel:
  connections:
    env1:
      redis:
        host: env1-logscache.redis.cache.windows.net
        port: 6380
        key:
      local:
        host: "127.0.0.1"
        port: 6379
    env2:
      redis:
        host: env2-logscache.redis.cache.windows.net
        port: 6380
        key:
      local:
        host: "127.0.0.1"
        port: 6378
```

### Indexes Clean Up

The old indexes are cleaned up by the [Curator](https://github.com/elastic/curator) tool which is executed daily by a cron job. Its configuration is available in [curator-actions.yaml](charts/kibana-logstash/templates/config/curator-actions.yaml) file. You should adjust it according to your needs.

### Index Templates

The [Elasticsearch Index Templates](https://www.elastic.co/guide/en/elasticsearch/reference/master/indices-templates.html) are installed automatically by a pre-install job. They are defined in the [pre-install-templates-config.yaml](charts/kibana-logstash/templates/pre-install-templates-config.yaml) file.

### Index Watches

The [Elasticsearch Watches](https://www.elastic.co/guide/en/elasticsearch/reference/master/watcher-api.html) are also installed automatically by a post-install job. They can be used to trigger any alert or notification based on search queries. The watches configuration is available in [post-install-watches-secret.yaml](charts/kibana-logstash/templates/post-install-watches-secret.yaml) file. You should update this configuration according to your needs.

### Elasticsearch License

In case you have an [Elasticsearch x-pack license](https://license.elastic.co/download), you can install it when [elasticsearch chart](charts/elasticsearch/README.md) is deployed.

## Installation

### NGINX Ingress Controller

The `nginx-ingress` will act as a frontend load balancer and it will provide TLS termination for the Kibana public endpoint. Get the latest version from [kubernetes/charts/stable/nginx-ingress](https://github.com/kubernetes/charts/tree/master/stable/nginx-ingress). Before starting the installation, updating e a few Helm values from `values.yaml` file is necessary.

Enable the Kubernetes RBAC by setting:

```console
rbac.create=true
```

And set the static public IP allocated in Azure, as a load balancer frontend IP:

```console
controller.service.loadBalancerIP: "<YOUR PUBLIC IP>"
```

Install now the helm package with the following commands:

```console
cd charts/stable/nginx-ingress
helm install -f values.yaml -n nginx-ingress .
```

After the installation is done, verify that the public IP is properly assigned to the controller.

```console
$> kubectl get svc nginx-ingress-nginx-ingress-controller

NAME                                     TYPE           CLUSTER-IP    EXTERNAL-IP        PORT(S)                      AGE
nginx-ingress-nginx-ingress-controller   LoadBalancer   10.0.26.141   <YOUR-PUBLIC-IP>   80:32321/TCP,443:31990/TCP   10m
```

### Elasticsearch Cluster

Kibana requires an Elasticsearch cluster which can be installed using the [elasticsearch chart](charts/elasticsearch/README.md). Create a deployment using the `deploy.sh` script available in the chart. Check the [README](charts/elasticsearch/README.md) file for more details:

```console
./deploy.sh -e acs -n elk
```

The command will install an Elasticsearch cluster in the `elk` namespace using the `acs` environment variables.

### Kibana and Logstash

You can install now the [kibana-logstash](charts/kibana-logstash) chart using the `deploy.sh` script available in the chart. Check the [README](charts/kibana-logstash/README.md) file for more details.

```console
./deploy.sh -n elk -d <DOMAIN> -v <KEYVAULT_NAME>
```

> Note to replace the `DOMAIN` with the Kibana DNS domain and the `KEYVAULT_NAME` with the Azure Key Vault name.

This command installs Kibana and Logstash in the `elk` namespace using the `acs` environment variables. If everything works well, the following output should be shown:

```console
Checking az command
Checking helm command
Checking kubectl command
Checking openssl command
Retrieving secrets from KeyVault:
  Fetching Kibana certificate and private key
  Fetching OAuth2 proxy secrets
  Fetching Elasticsearch Watcher secrets
  Fetching 'acs' environemnt specific secrets
Installing mse-elk helm chart
Done
```

And the deployment should look like this:

```console
$> kubectl get pods --namespace elk

NAME                                                           READY     STATUS    RESTARTS   AGE
es-client-7b8599646c-7qfpr                                     1/1       Running   0          1h
es-client-7b8599646c-nkqjb                                     1/1       Running   0          1h
es-client-7b8599646c-x2qjr                                     1/1       Running   0          1h
es-data-0                                                      1/1       Running   0          1h
es-data-1                                                      1/1       Running   0          1h
es-data-2                                                      1/1       Running   0          1h
es-data-3                                                      1/1       Running   0          1h
es-data-4                                                      1/1       Running   0          1h
es-data-5                                                      1/1       Running   0          1h
es-master-57d94ff4f7-ncfzg                                     1/1       Running   0          1h
es-master-57d94ff4f7-sfpk7                                     1/1       Running   0          1h
es-master-57d94ff4f7-szrqb                                     1/1       Running   0          1h
kibana-56c7bf5c46-mwxtl                                        2/2       Running   0          10m
kibana-56c7bf5c46-t862s                                        2/2       Running   0          10m
kibana-56c7bf5c46-tdmf8                                        2/2       Running   0          10m
logstash-0                                                     2/2       Running   0          10m
logstash-1                                                     2/2       Running   0          10m
logstash-2                                                     2/2       Running   0          10m
nginx-ingress-nginx-ingress-controller-7f7488c7c7-wkx42        1/1       Running   0          1h
nginx-ingress-nginx-ingress-default-backend-7c8bbc9879-cvl79   1/1       Running   0          1h
```

### Rolling Update

You can upgrade the charts after the initial installation whenever you have a change, by simply executing again the deployment scripts with the same arguments. Helm will create a new release for you.

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit [https://cla.microsoft.com](https://cla.microsoft.com).

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
