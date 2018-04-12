# elasticsnapshot

This is a tool which can be used to create and to restore a snapshot of an entire Elasticsearch cluster.

## Installation

```bash
go get github.com/Azure/helm-elasticstack/tools/elasticsnapshot
```

Alternatively you can build the docker image by cloning the repository and executing the following command:

```bash
make image

```

## Usage

``` 
$> elasticsnapshot help
Usage: elasticsnapshot <flags> <subcommand> <subcommand args>

Subcommands:
        commands         list all command names
        create           create a new snapshot of the entire Elasticsearch cluster in an Azure storage
        flags            describe all known top-level flags
        help             describe subcommands and their syntax
        restore          restore an entire Elasticsearch cluster snapshot from Azure storage
        status           retrieves the status of an Elasticsearch snapshot


Use "elasticsnapshot flags" for a list of top-level flags

```

## Create Snapshot

You can define the basic authentication credentials used by your Elasticsearch cluster in a `auth-file.json` as follows:

```json
{
  "username": "<USER NAME>",
  "password": "<PASSWORD>"
}

```
A snapshot of the Elasticsearch cluster can be created in Azure storage with the following command:

```bash
elasticsnapshot create -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> \
-auth-file=auth-file.json -respository <REPOSITORY-NAME> -snapshot <SNAPSHOT-NAME>
```

or run the tool in a docker container:

```bash
docker run --rm -v ${PWD}:/config -t mseoss/elasticsnapshot create -host=<ELASTICSEARCH-HOST> \
-port=<ELASTICSEARCH-PORT> -auth-file=/config/auth-file.json -respository <REPOSITORY-NAME> -snapshot <SNAPSHOT-NAME>
```

The snapshot will be created asynchronously. You can retrieve the status of the snapshot in order to check if the snapshot is being created.

```bash
elasticsnapshot status -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> \
-auth-file=auth-file.json -respository <REPOSITORY-NAME> -snapshot <SNAPSHOT-NAME>
```

## Restore Snapshot

A snapshot can be restored from Azure storage as follows:

```bash
elasticsnapshot restore -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> \
-auth-file=auth-file.json -respository <REPOSITORY-NAME> -snapshot <SNAPSHOT-NAME>
```

The snapshot is restored asynchronously. You can check the status of the restore operations by running the command:

```bash
elasticsnapshot status -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> \
-auth-file=auth-file.json -respository <REPOSITORY-NAME> -snapshot <SNAPSHOT-NAME>
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
