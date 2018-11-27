# Introduction

Docker image based on official [Elasticsearch image](https://www.elastic.co/guide/en/logstash/current/docker.html) with some extra plugins installed.

Plugins:

* `repository-azure` - [Azure Repository Plugin](https://www.elastic.co/guide/en/elasticsearch/plugins/current/repository-azure.html#repository-azure)

## Build

The `VERSION` environment variable defines the version of the elasticsearch base image.

```bash
export VERSION=6.4.3
docker build -t mseoss/elasticsearch:${VERSION} --build-arg VERSION=${VERSION} .
docker push mseoss/elasticsearch:${VERSION}
```

## Run

Elasticsearch image can be started for testing with the following command:

```bash
docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" mseoss/elasticsearch:${VERSION}
```
