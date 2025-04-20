{{/*
Expand the name of the chart.
*/}}
{{- define "kintegrity.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kintegrity.fullname" -}}
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
{{- define "kintegrity.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kintegrity.labels" -}}
helm.sh/chart: {{ include "kintegrity.chart" . }}
{{ include "kintegrity.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kintegrity.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kintegrity.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kintegrity.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kintegrity.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the secret digests mapping to use
*/}}
{{- define "kintegrity.digestsSecretName" -}}
{{- default (printf "%s-%s" (include "kintegrity.fullname" .) "digests-mapping") .Values.digestsMapping.secretName }}
{{- end }}

{{/*
Create the name of the webhook cert secret to use
*/}}
{{- define "kintegrity.webhookSecretName" -}}
{{- default (printf "%s-%s" (include "kintegrity.fullname" .) "webhook-cert") .Values.certificates.webhook.secretName }}
{{- end }}

{{/*
Create the name of the metrics cert secret to use
*/}}
{{- define "kintegrity.metricsSecretName" -}}
{{- default (printf "%s-%s" (include "kintegrity.fullname" .) "metrics-cert") .Values.certificates.metrics.secretName }}
{{- end }}
