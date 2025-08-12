{{- /*
Common labels for all resources.
This is a good practice to ensure consistency across the entire chart.
*/}}
{{- define "km-kube-agent.labels" -}}
helm.sh/chart: {{ include "km-kube-agent.chart" . }}
{{- with .Values.podLabels }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{- /*
Selector labels for the various workloads.
These labels are used in Deployment and DaemonSet selectors.
*/}}
{{- define "km-kube-agent.daemonset.selectorLabels" -}}
{{- toYaml .Values.daemonsetLabels }}
{{- end }}

{{- define "km-kube-agent.deployment.selectorLabels" -}}
{{- toYaml .Values.deploymentLabels }}
{{- end }}

{{- define "km-kube-agent.configUpdater.selectorLabels" -}}
{{- toYaml .Values.configUpdaterLabels }}
{{- end }}

{{- /*
Create a default fully qualified app name.
We truncate it at 63 chars because of K8s name restrictions.
*/}}
{{- define "km-kube-agent.fullname" -}}
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

{{- /*
Create the name of the chart.
*/}}
{{- define "km-kube-agent.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}
