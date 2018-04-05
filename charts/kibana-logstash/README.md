## Introduction

This chart bootstraps [Kibana](https://www.elastic.co/guide/en/kibana/current/index.html) with [Logstash](https://www.elastic.co/guide/en/logstash/current/index.html) on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites
 - Kubernetes 1.8+ (e.g. deployed with [Azure Container Service (AKS)](https://docs.microsoft.com/en-us/azure/aks/intro-kubernetes))

## Configuration

The following table lists some of the configurable parameters of the `kibana-logstash` chart and their default values:

| Parameter                                      | Description                                                         | Default                                                                |
| ---------------------------------------------- | ------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| `image.pullPolicy`                             | General image pull policy                                           | `Always`                                                               |
| `image.pullSecrets`                            | General image pull secrets                                          | `nil` (does not add image pull secrets to deployed pods)               |
| `kibana.image.repository`                      | Kibana image                                                        | `docker.elastic.co/kibana/kibana`                                      |
| `kibana.image.tag`                             | Kibana image tag                                                    | `6.2.3`                                                                |
| `kibana.replicas`                              | Number of Kibana instances started                                  | `3`                                                                    |
| `kibana.ingress.host`                          | Kibana DNS domain                                                   | `nil` (must be provided during installation)                           |
| `kibana.ingress.public.cert`                   | Kibana public TLS certificate                                       | `nil` (must be provided during installation)                           |
| `kibana.ingress.private.key`                   | Kibana private TLS key                                              | `nil` (must be provided during installation)                           |
| `logstash.image.repository`                    | Logstash image                                                      | `mseoss/logstash`                                                      |
| `logstah.image.tag`                            | Logstash image tag                                                  | `6.2.3`                                                                |
| `logstah.replicas`                             | Number of Logtash instances started                                 | `3`                                                                    |
| `logstah.queue.storageclass`                   | Storage class used for Logstash queue PV                            | `default`                                                              |
| `logstah.queue.distk_capacity`                 | Disk capacity of Logstash queue PV                                  | `50Gi`                                                                 |
| `stunnel.image.repository`                     | Stunnel image                                                       | `mseoss/stunnel`                                                       |
| `stunnel.image.tag`                            | Stunnel image tag                                                   | `5.44`                                                                 |
| `stunnel.connections.dev.redis.host`           | Address of Redis where the logs for `dev` environment are cached    | `dev-logschache.redis.cache.windows.net`                               |
| `stunnel.connections.dev.redis.port`           | Port of Redis where logs for `dev` environment are cached           | `6380`                                                                 |
| `stunnel.connections.dev.redis.key`            | Key of Redis where logs for `dev` environment are cached            | `nil` (must be provided during installation)                           |
| `stunnel.connections.dev.logcal.host`          | Local host where Redis connection for `dev` environment is tunneled | `127.0.0.1`                                                            |
| `stunnel.connections.dev.logcal.port`          | Local port where Redis connection for `dev` environment is tunneled | `6379`                                                                 |
| `oauth.image.repository`                       | oauth2_proxy image                                                  | `mseoss/oauth2_proxy`                                                  |
| `oauth.image.tag`                              | oauth2_proxy image tag                                              | `v2.2`                                                                 |
| `oauth.client.id`                              | Azure AD application ID                                             | `nil` (must be provided during installation)                           |
| `oauth.client.secret`                          | Azure AD application secret                                         | `nil` (must be provided during installation)                           |
| `oauth.cookie.secret`                          | Secrete used to sign the Kibana SSO cookie                          | `nil` (must be provided during installation)                           |
| `oauth.cookie.expire`                          | Kibana SSO cookie expiration time                                   | `168h0m`                                                               |
| `oauth.cookie.refresh`                         | Kibana SSO cookie refresh time                                      | `60m`                                                                  |
| `curator.image.repository`                     | Curator image                                                       | `docker.io/bobrik/curator`                                             |
| `curator.image.tag`                            | Curator image tag                                                   | `latest`                                                               |
| `curator.install`                              | Indicates if curator cron job is created                            | `true`                                                                 |
| `curator.image.index_prefix`                   | Prefix of the index over which curator runs                         | `dev` (should be the same like stunnel.connection.[env]                |
| `templates.image.repository`                   | Elastic template tool image                                         | `mseoss/elastictemplate`                                               |
| `templates.image.tag`                          | Elastic template image tag                                          | `latest`                                                               |
| `templates.image.install`                      | Indicates if elastic template pre-install job is executed           | `true`                                                                 |
| `watcher.image.repository`                     | Elastic watcher tool image                                          | `mseoss/elasticwatcher`                                                |
| `watcher.image.tag`                            | Elastic watcher image tag                                           | `latest`                                                               |
| `watcher.image.install`                        | Indicates if elastic watcher post-install job is executed           | `true`                                                                 |
| `watcher.webhooks.teams`                       | Microsoft teams webhook (watcher will post here the alerts)         | `nil` (must be provided during installation)                           |
| `watcher.indices`                              | Index prefixes where watches will be executed                       | ``"dev-logstash-*\"`(env prefix the same like stunnel.connection.[env] |

> Note that you can define multiple Redis connections. The helm chart will create a Logstash data pipeline for each of connection.

## Installing the Chart

The chart can be installed with the `deploy.sh` script. There are a few arguments which should be provided as input:
- The environment which contains the helm values (default is `acs`)
- A namespace (defualt is `elk`)
- The public DNS domain used by Kibana
- The name of the Azure KeyVault where the secrets are stored

```console
./deploy.sh -e acs -n elk -d my.kibana.domain.com -v keyvault-name
```

## Uninstalling the Chart

The chart can be uninstalled/deleted as follows:

```console
helm delete --purge kibana-logstash
```

This command removes all the Kubernetes resources associated with the chart and deletes the helm release.


## Validate the Chart

### Lint

You can validate that the chart has not lint warnings during development.

```console
helm lint -f environments/acs/values.yaml
```

### Template rendering

You can validate if the chart is properly rendered using the `helm template` command. A `dry run mode` is built into the deployment script. You just need to execute the script with the `-t` option:

```console
./deploy.sh -t -n elk
```

## Scale up the Logstash nodes

The logstash nodes can be easily scaled up/down with the following command:

```console
kubectl scale --namespace elk statefulset/logstash --replicas 6
```
