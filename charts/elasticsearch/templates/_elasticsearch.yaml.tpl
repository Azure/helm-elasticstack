cluster:
  name: ${CLUSTER_NAME}

node:
  master: ${NODE_MASTER}
  name: ${NODE_NAME}
  data: ${NODE_DATA}
  ingest: ${NODE_INGEST}
  max_local_storage_nodes: ${MAX_LOCAL_STORAGE_NODES}

network:
  host: ${NETWORK_HOST}

path:
  data: /data/data
  logs: /data/log

bootstrap:
  memory_lock: ${MEMORY_LOCK}

http:
  enabled: ${HTTP_ENABLE}
  compression: true
  cors:
    enabled: ${HTTP_CORS_ENABLE}
    allow-origin: ${HTTP_CORS_ALLOW_ORIGIN}

discovery:
  zen:
    ping.unicast.hosts: ${DISCOVERY_SERVICE}
    minimum_master_nodes: ${MINIMUM_MASTER_NODES}

xpack:
  security:
    enabled: {{ .Values.xpack.security | default false }}
  monitoring:
    enabled: {{ .Values.xpack.monitoring | default false }}