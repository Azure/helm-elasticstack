# Introduction

Docker image for [stunnel](https://www.stunnel.org) TLS proxy.

# Build

The `VERSION` environment variable defines the version of the stunnel package.

```
export VERSION=5.44
docker build -t mseoss/stunnel:${VERSION} --build-arg VERSION=${VERSION} .
docker push mseoss/stunnel:${VERSION}
```

# Run

stunnel can be started in a docker container with the following command:

```
docker run mseoss/stunnel:${VERSION}
```
