ARG VERSION=latest
FROM docker.elastic.co/elasticsearch/elasticsearch:${VERSION}

# Install Plugins
RUN elasticsearch-plugin install repository-azure --verbose --batch
