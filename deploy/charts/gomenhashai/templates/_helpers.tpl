{{/*
Expand the name of the chart.
*/}}
{{- define "gomenhashai.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "gomenhashai.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "gomenhashai.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "gomenhashai.labels" -}}
helm.sh/chart: {{ include "gomenhashai.chart" . }}
{{ include "gomenhashai.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "gomenhashai.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gomenhashai.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "gomenhashai.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "gomenhashai.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the secret digests mapping to use
*/}}
{{- define "gomenhashai.digestsSecretName" -}}
{{- default (printf "%s-%s" (include "gomenhashai.fullname" .) "digests-mapping") .Values.digestsMapping.secretName }}
{{- end }}

{{/*
Create the name of the webhook cert secret to use
*/}}
{{- define "gomenhashai.webhookSecretName" -}}
{{- default (printf "%s-%s" (include "gomenhashai.fullname" .) "webhook-cert") .Values.certificates.webhook.secretName }}
{{- end }}

{{/*
Create the name of the metrics cert secret to use
*/}}
{{- define "gomenhashai.metricsSecretName" -}}
{{- default (printf "%s-%s" (include "gomenhashai.fullname" .) "metrics-cert") .Values.certificates.metrics.secretName }}
{{- end }}

{{/*
Generate self signed CA and certificate/key pair
*/}}
{{- if not (or (index .Values "certificates" "cert-manager" "enabled") .Values.certificates.webhook.secretName .Values.certificates.metrics.secretName) }}
{{- define "gomenhashai.ca" -}}
{{- genCA "gomenhashai-ca" .Values.certificates.duration }}
{{- end }}
{{- define "gomenhashai.webhookcert" -}}
{{- genSignedCert (printf "%s--webhook-service.%s.svc" (include "gomenhashai.fullname" .) .Release.Namespace) nil nil .Values.certificates.duration (include "gomenhashai.ca" .) }}
{{- end }}
{{- define "gomenhashai.metricscert" -}}
{{- genSignedCert (printf "%s--metrics-service.%s.svc" (include "gomenhashai.fullname" .) .Release.Namespace) nil nil .Values.certificates.duration (include "gomenhashai.ca" .) }}
{{- end }}
{{- end }}
