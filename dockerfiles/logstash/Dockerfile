ARG VERSION=latest
FROM docker.elastic.co/logstash/logstash:${VERSION}

# Install Plugins
RUN logstash-plugin install logstash-output-csv
RUN logstash-plugin install logstash-output-elasticsearch
RUN logstash-plugin install logstash-input-redis
