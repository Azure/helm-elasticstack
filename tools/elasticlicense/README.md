# elasticlicense

This is a tool which can be used to install the x-pack license into an Elasticsearch cluster.

## Installation

```bash
go get github.com/Azure/helm-elasticstack/tools/elasticlicense
```

Alternatively you can build the docker image by cloning the repository and executing the following command:

```bash
make image
```

## Install a new license

Download the license form [Elasticsearch support](https://license.elastic.co/download) and store it into a `license.json` file.

You should also define the basic authentication credentials used by your Elasticsearch cluster in a `auth-file.json` as follows:

```json
{
  "username": "<USER NAME>",
  "password": "<PASSWORD>"
}

```
The license can be installed by executing the command:

```bash
elasticlicense install -license-file=license.json -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=auth-file.json
```

or run the tool in a docker container:

```bash
docker run --rm -v ${PWD}:/config -t docker.io/mse/elasticlicense install -license-file=/config/license.json \
-host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=/config/auth-file.json
```

The installed license can be viewed with the following command:

```bash
elasticlicense view -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=auth-file.json
```

## Development

You can execute the tests and build the tool using the default make target:

```bash
make
```

To build and publish the docker image execute:

```bash
make image
make image-push
```
