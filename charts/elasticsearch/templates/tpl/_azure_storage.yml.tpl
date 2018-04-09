{{- if .Values.cloud -}}
{{- if .Values.cloud.azure }}
cloud:
  azure:
    storage:
      storage_account:
        account: {{- .Values.cloud.azure.storage.account }}
        key: {{- .Values.cloud.azure.storage.key }}
{{- end }}
{{- end -}}
