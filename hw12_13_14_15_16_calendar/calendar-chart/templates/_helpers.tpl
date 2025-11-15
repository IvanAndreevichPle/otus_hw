{{/*
Expand the name of the chart.
*/}}
{{- define "calendar-chart.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "calendar-chart.fullname" -}}
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
{{- define "calendar-chart.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "calendar-chart.labels" -}}
helm.sh/chart: {{ include "calendar-chart.chart" . }}
{{ include "calendar-chart.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "calendar-chart.selectorLabels" -}}
app.kubernetes.io/name: {{ include "calendar-chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Calendar API labels
*/}}
{{- define "calendar-chart.calendar.labels" -}}
helm.sh/chart: {{ include "calendar-chart.chart" . }}
app.kubernetes.io/name: {{ include "calendar-chart.name" . }}-calendar
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: calendar-api
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Calendar API selector labels
*/}}
{{- define "calendar-chart.calendar.selectorLabels" -}}
app.kubernetes.io/name: {{ include "calendar-chart.name" . }}-calendar
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: calendar-api
{{- end }}

{{/*
Scheduler labels
*/}}
{{- define "calendar-chart.scheduler.labels" -}}
helm.sh/chart: {{ include "calendar-chart.chart" . }}
app.kubernetes.io/name: {{ include "calendar-chart.name" . }}-scheduler
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: scheduler
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Scheduler selector labels
*/}}
{{- define "calendar-chart.scheduler.selectorLabels" -}}
app.kubernetes.io/name: {{ include "calendar-chart.name" . }}-scheduler
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: scheduler
{{- end }}

{{/*
Sender labels
*/}}
{{- define "calendar-chart.sender.labels" -}}
helm.sh/chart: {{ include "calendar-chart.chart" . }}
app.kubernetes.io/name: {{ include "calendar-chart.name" . }}-sender
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: sender
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Sender selector labels
*/}}
{{- define "calendar-chart.sender.selectorLabels" -}}
app.kubernetes.io/name: {{ include "calendar-chart.name" . }}-sender
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: sender
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "calendar-chart.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "calendar-chart.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

