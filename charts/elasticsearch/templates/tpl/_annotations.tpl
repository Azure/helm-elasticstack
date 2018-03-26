{{- define "annotations" -}}
{{ $scope_var := . }}
        config/checksum: {{ print $scope_var | sha256sum }}
{{- end -}}