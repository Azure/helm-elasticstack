# elasticwatcher

This is a tool which can be used to manage the watches in [Elasticsarch Watcher](https://www.elastic.co/guide/en/elasticsearch/reference/5.6/watcher-api.html).

## Installation

```bash
go get github.com/Azure/helm-elasticstack/tools/elasticwatcher
```

Alternatively you can build the docker image by cloning the repository and executing the following command:

```bash
make image
```

## Usage 

```
$> elasticwatcher help
Usage: elasticwatcher <flags> <subcommand> <subcommand args>

Subcommands:
        activate         Activate a list of watches from Elasicsearch Watcher
        commands         list all command names
        create           Register a list of watches in Elasicsearch Watcher or update them
        deactivate       Deactivate a list of watches from Elasicsearch Watcher
        delete           Delete a list of watches from Elasicsearch Watcher
        flags            describe all known top-level flags
        help             describe subcommands and their syntax
        list             List all watches installed in Elasticsearch Watcher
        retrieve         Retrieve a list of watches from Elasicsearch Watcher by their name


Use "elasticwatcher flags" for a list of top-level flags

```

You can define the basic authentication credentials used by your Elasticsearch cluster in a `auth-file.json` as follows:

```json
{
  "username": "<USER NAME>",
  "password": "<PASSWORD>"
}

```

The watches can be defined in a `watches.json`, where you have to specified the name of the watch and its body:

```json
{
   "watches": [
       {
           "name": "watch_name"
           "body": {
              "trigger" : {
                  "schedule" : { "cron" : "0 0/1 * * * ?" }
                },
                "input" : {
                  "search" : {
                    "request" : {
                      "indices" : [
                        "logstash*"
                      ],
                      "body" : {
                        "query" : {
                          "bool" : {
                            "must" : {
                              "match": {
                                "response": 404
                              }
                            },
                            "filter" : {
                              "range": {
                                "@timestamp": {
                                  "from": "{{ctx.trigger.scheduled_time}}||-5m",
                                  "to": "{{ctx.trigger.triggered_time}}"
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                },
                "condition" : {
                  "compare" : { "ctx.payload.hits.total" : { "gt" : 0 }}
                },
                "actions" : {
                  "email_admin" : {
                    "email" : {
                      "to" : "admin@domain.host.com",
                      "subject" : "404 recently encountered"
                    }
                  }
                }
           }
       }
   ]

}
```

The body contains the effective watch definition and it should be defined according with the Elasticsearch's [guidelines](https://www.elastic.co/guide/en/x-pack/6.1/how-watcher-works.html#watch-active-state).

The watches can be created executing the command:

```bash
elasticwatcher create -watches-file=watches.json -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=auth-file.json
```

or using a docker container:

```bash
docker run --rm -v ${PWD}:/config -t docker.io/mse/elasticwatcher create -watches-file=/config/watches.json \
-host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=/config/auth-file.json
```

The created watches can be retrieved with:

```bash
elasticwatcher retrieve -watches=watch-name1,watch-name2 -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=auth-file.json
```

Also you can delete a list of watches as follows:

```bash
elasticwatcher delete -watches=watch-name1,watch-name2 -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=auth-file.json
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
