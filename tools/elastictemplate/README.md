# elastictemplate

This is a tool which can be used to manage the [Elasticsarch Index Templates](https://www.elastic.co/guide/en/elasticsearch/reference/5.4/indices-templates.html).

## Installation

```bash
go get github.com/Azure/helm-elasticstack/tools/elastictemplate
```

Alternatively you can build the docker image by cloning the repository and executing the following command:

```bash
make image
```

## Usage

```
./elastictemplate -h
Usage: elastictemplate <flags> <subcommand> <subcommand args>

Subcommands:
        commands         list all command names
        create           Create an Elasticsearch Index Template or update an existing one
        delete           Delete the templates from Elasicsearch
        flags            describe all known top-level flags
        help             describe subcommands and their syntax
        list             List all Elasticsearch Index Templates
        retrieve         Retrieve the content of Elasicsearch Index Templates


Use "elastictemplate flags" for a list of top-level flags

```

You can define the basic authentication credentials used by your Elasticsearch cluster in a `auth-file.json` as follows:

```json
{
  "username": "<USER NAME>",
  "password": "<PASSWORD>"
}

```

The templates can be defined in a `templates.json`, where you have to specified the name of the template and its body:

```json
{
   "templates": [
       {
           "name": "template_1"
           "body": {
            "template" : "te*",
                "settings" : {
                    "number_of_shards" : 1
                },
                "aliases" : {
                    "alias1" : {},
                    "alias2" : {
                        "filter" : {
                            "term" : {"user" : "test" }
                        },
                        "routing" : "test"
                    },
                    "{index}-alias" : {}
                }
           }
       }
   ]

}
```

The body contains the definition of the index template, and it should be created according with the Elasticsearch's [guidelines](https://www.elastic.co/guide/en/elasticsearch/reference/5.4/indices-templates.html).

The templates can be created/updated by executing the command:

```bash
elstictemplate create -templates-file=templates.json -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=auth-file.json
```

or using a docker container:

```bash
docker run --rm -v ${PWD}:/config -t docker.io/mse/elastictemplate create -templates-file=/config/templates.json \
-host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=/config/auth-file.json
```

The existing templates can be listed with:

```bash
elastictemplate list -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=auth-file.json
```

Or you can delete some existing templates as follows:

```bash
elstictemplate delete -templates=template-name1,template-name2 -host=<ELASTICSEARCH-HOST> -port=<ELASTICSEARCH-PORT> -auth-file=auth-file.json
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
