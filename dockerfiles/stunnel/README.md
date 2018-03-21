# Introduction

Docker image for [stunnel](https://www.stunnel.org) TLS proxy.

# Build

The `VERSION` environment variable defines the version of the stunnel package.

```
export VERSION=5.44
docker build -t docker.io/mse/stunnel:${VERSION} --build-arg VERSION=${VERSION} .
docker push docker.io/mse/stunnel:${VERSION}
```

# Run

stunnel can be started in a docker container with the following command:

```
docker run docker.io/mse/stunnel:${VERSION}
```
