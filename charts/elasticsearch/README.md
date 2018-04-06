## Introduction

This chart bootstraps an [Elasticsearch cluster](https://www.elastic.co/guide/en/elasticsearch/reference/current/docker.html) on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

It is based on [clockworksoul/helm-elasticsearch](https://github.com/clockworksoul/helm-elasticsearch) chart.

## Prerequisites
 - Kubernetes 1.8+ e.g. deployed with [Azure Container Service (AKS)](https://docs.microsoft.com/en-us/azure/aks/intro-kubernetes)

## Configuration

The following table lists some of the configurable parameters of the `elasticsearch` chart and their default values:

| Parameter                                      | Description                              | Default                                                   |
| ---------------------------------------------- | ---------------------------------------- | --------------------------------------------------------- |
| `image.repository`                             | elasticsearch image                      | `docker.elastic.co/elasticsearch/elasticsearch`           |
| `image.tag`                                    | specify image tag                        | `6.2.3`                                                   |
| `image.pullSecrets`                            | specify image pull secrets               | `nil` (does not add image pull secrets to deployed pods)  |
| `image.pullPolicy`                             | image pull policy                        | `Always`                                                  |
| `license.install`                              | install x-pack license                   | `false`                                                   |
| `license.image.repository`                     | elasticlicense tool image                | `mseoss/elasticlicense`                         |
| `license.image.tag`                            | elasticlicense tool image tag            | `latest`                                                  |
| `license.image.pullPolicy`                     | elasticlicense tool image pull policy    | `Always`                                                  |
| `license.value`                                | x-pack license                           | `nil` (must be provided during installation)              |
| `es_master.replicas`                           | number of master nodes started           | `3`                                                       |
| `es_client.replicas`                           | number of ingest nodes started           | `3`                                                       |
| `es_data.replicas`                             | number of data nodes started             | `3`                                                       |
| `es_data.disk_capacity`                        | size of the disk attacked to a data node | `100Gi`                                                   |
| `es_data.storageclass`                         | data node disk storage class             | `managed-premium`                                         |
| `es_master/es_client/es_data.env.ES_JAVA_OPTS` | Java VM options                          | `-Xms1g -Xmx1g`                                           |

## Installing the Chart

The chart can be installed with the `deploy.sh` script. An environment and namespace can be provided as input. The script will use by default the `acs` environment and `elk` namespace.

```console
./deploy.sh -e acs -n elk
```

Alternatively you can also installed automatically the [Elasticsearch x-pack license](https://license.elastic.co/download) after the deployment. First you need to activate the 
license installation in [Helm values](charts/elasticserch/environments/acs/values.yaml) by setting the `license.install=true` and you may also want to enable the `x-pack features` in [Elasticsearch config](charts/elasticsearch/templates/config.config.yaml).

```console
./deploy.sh -e acs -n elk -l license.json
```

## Uninstalling the Chart

The chart can be uninstalled/deleted as follows:

```console
helm delete --purge elasticsearch
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

## Verify the health of the Elasticsearch cluster

The health of the Elasticsearch cluster can be checked after the deployment from any of the mater nodes.

Open a shell into one of the master nodes:

```console
 kubectl exec -ti --namespace elk es-master -- /bin/bash
```

You can get now the health status directly from Elasticsearch API:

```console
curl http://localhost:9200/_cluster/health?pretty
{
  "cluster_name" : "es-cluster",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 9,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 7,
  "active_shards" : 14,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}
```

## Scale up the data nodes

The data nodes can be easily scaled up/down with the following command:

```console
kubectl scale --namespace elk statefulset/es-data --replicas 10
```

## Access the Elasticsearch API

The Elasticsearch API is only exposed to the Kubernetes cluster/namespace and it can be accessed at http://elasticsearch:9200 address from any POD running in the same namespace.
