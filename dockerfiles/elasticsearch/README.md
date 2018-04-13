# Introduction

Docker image based on official [Elasticsearch image](https://www.elastic.co/guide/en/logstash/current/docker.html) with some extra plugins installed.

Plugins:
* repository-azure 

# Build

The `VERSION` environment variable defines the version of the elasticsearch base image.

```
export VERSION=6.2.3
docker build -t mseoss/elasticsearch:${VERSION} --build-arg VERSION=${VERSION} .
docker push mseoss/elasticsearch:${VERSION}
```

# Run


Elasticsearch image can be started for testing with the following command:

```
docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" mseoss/elasticsearch:${VERSION}
```

