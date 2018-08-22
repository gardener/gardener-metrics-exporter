{{- define "rbacversion" -}}
rbac.authorization.k8s.io/v1
{{- end -}}

{{- define "deploymentversion" -}}
{{- if semverCompare ">= 1.9" .Capabilities.KubeVersion.GitVersion -}}
apps/v1
{{- else -}}
apps/v1beta2
{{- end -}}
{{- end -}}