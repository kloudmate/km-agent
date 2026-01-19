# Supported OpenTelemetry Components

This document lists all the OpenTelemetry Collector components that are currently supported by the KloudMate Agent.

## Table of Contents
- [Extensions](#extensions)
- [Receivers](#receivers)
- [Processors](#processors)
- [Exporters](#exporters)
- [Connectors](#connectors)

---

## Extensions

Extensions provide capabilities that can be added to the collector, but which do not require direct access to telemetry data.

| Extension | Description | Use Case | Documentation |
|-----------|-------------|----------|---------------|
| **Memory Limiter** | Monitors and controls memory usage | Prevents out-of-memory errors by applying backpressure when memory limits are reached | [Docs](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/memorylimiterprocessor) |
| **Z-Pages** | Provides in-process web pages for diagnostics | Offers debugging endpoints for live troubleshooting and performance analysis | [Docs](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/zpagesextension) |
| **File Storage** | Provides persistent storage capabilities | Used by receivers/processors that need to maintain state across restarts | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/storage/filestorage) |
| **Health Check** | Exposes health check endpoints | Enables Kubernetes liveness and readiness probes for the collector | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/healthcheckextension) |

---

## Receivers

Receivers are responsible for getting data into the collector. A receiver can be push or pull-based.

### Infrastructure & Host Monitoring

| Receiver | Description | Metrics Collected | Documentation |
|----------|-------------|-------------------|---------------|
| **OTLP** | Receives data via OTLP protocol | Metrics, traces, and logs from OTLP-compatible sources | [Docs](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/otlpreceiver) |
| **Host Metrics** | Collects host-level metrics | CPU, memory, disk, network, filesystem metrics | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) |
| **Docker Stats** | Collects Docker container metrics | Container CPU, memory, network, and block I/O statistics | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/dockerstatsreceiver) |
| **Kubelet Stats** | Collects metrics from Kubernetes Kubelet | Pod and container resource usage metrics | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kubeletstatsreceiver) |

### Kubernetes Monitoring

| Receiver | Description | Metrics Collected | Documentation |
|----------|-------------|-------------------|---------------|
| **K8s Cluster** | Collects Kubernetes cluster-level metrics | Node, pod, deployment, and service metrics | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8sclusterreceiver) |
| **K8s Objects** | Watches Kubernetes objects and converts them to logs | Events, pod status changes, deployment updates | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8sobjectsreceiver) |
| **K8s Events** | Collects Kubernetes events | Cluster events as log entries | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8seventsreceiver) |

### Database Receivers

| Receiver | Description | Database Type | Documentation |
|----------|-------------|---------------|---------------|
| **MySQL** | Collects metrics from MySQL databases | MySQL 5.7+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mysqlreceiver) |
| **PostgreSQL** | Collects metrics from PostgreSQL databases | PostgreSQL 9.6+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/postgresqlreceiver) |
| **MongoDB** | Collects metrics from MongoDB instances | MongoDB 4.0+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mongodbreceiver) |
| **MongoDB Atlas** | Collects metrics from MongoDB Atlas | MongoDB Atlas cloud service | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mongodbatlasreceiver) |
| **Redis** | Collects metrics from Redis instances | Redis 5.0+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/redisreceiver) |
| **Oracle DB** | Collects metrics from Oracle databases | Oracle Database | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/oracledbreceiver) |
| **SQL Server** | Collects metrics from Microsoft SQL Server | SQL Server 2012+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/sqlserverreceiver) |
| **SAP HANA** | Collects metrics from SAP HANA databases | SAP HANA | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/saphanareceiver) |
| **Elasticsearch** | Collects metrics from Elasticsearch clusters | Elasticsearch 7.x+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/elasticsearchreceiver) |

### Web Server & Application Receivers

| Receiver | Description | Supported Versions | Documentation |
|----------|-------------|--------------------|---------------|
| **Apache** | Collects metrics from Apache HTTP Server | Apache 2.4+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/apachereceiver) |
| **Nginx** | Collects metrics from Nginx servers | Nginx with stub_status | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/nginxreceiver) |
| **IIS** | Collects metrics from Microsoft IIS | Windows Server 2016+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/iisreceiver) |
| **RabbitMQ** | Collects metrics from RabbitMQ message broker | RabbitMQ 3.8+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/rabbitmqreceiver) |
| **Kafka Metrics** | Collects metrics from Apache Kafka | Kafka 2.0+ | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kafkametricsreceiver) |

### Log Receivers

| Receiver | Description | Log Sources | Documentation |
|----------|-------------|-------------|---------------|
| **Filelog** | Tails and parses log files | Any file-based logs with customizable parsing | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) |
| **Journald** | Collects logs from systemd journal | Linux systemd journal logs | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/journaldreceiver) |
| **Syslog** | Receives syslog messages | RFC3164 and RFC5424 syslog formats | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/syslogreceiver) |
| **Fluent Forward** | Receives logs via Fluentd forward protocol | Fluentd/Fluent Bit sources | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/fluentforwardreceiver) |

### Cloud Platform Receivers

| Receiver | Description | Cloud Provider | Documentation |
|----------|-------------|----------------|---------------|
| **AWS CloudWatch** | Collects logs from AWS CloudWatch Logs | AWS | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/awscloudwatchreceiver) |
| **AWS CloudWatch Metrics** | Collects metrics from AWS CloudWatch | AWS | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/awscloudwatchmetricsreceiver) |
| **AWS Container Insights** | Collects container insights from ECS/EKS | AWS ECS/EKS | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/awscontainerinsightreceiver) |
| **AWS ECS Container Metrics** | Collects metrics from ECS containers | AWS ECS | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/awsecscontainermetricsreceiver) |
| **Azure Monitor** | Collects metrics from Azure Monitor | Microsoft Azure | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/azuremonitorreceiver) |
| **Google Cloud Monitoring** | Collects metrics from Google Cloud Monitoring | Google Cloud Platform | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/googlecloudmonitoringreceiver) |

### Additional Receivers

| Receiver | Description | Use Case | Documentation |
|----------|-------------|----------|---------------|
| **Prometheus** | Scrapes Prometheus metrics endpoints | Applications exposing Prometheus metrics | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver) |
| **HTTP Check** | Performs HTTP health checks | Synthetic monitoring and endpoint availability | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/httpcheckreceiver) |
| **SQL Query** | Executes custom SQL queries for metrics | Custom database metrics extraction | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/sqlqueryreceiver) |
| **Netflow** | Collects network flow data | Network traffic analysis | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/netflowreceiver) |
| **vCenter** | Collects metrics from VMware vCenter | VMware infrastructure monitoring | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/vcenterreceiver) |

---

## Processors

Processors are run on data between being received and being exported. Processors are optional.

### Essential Processors

| Processor | Description | Use Case | Documentation |
|-----------|-------------|----------|---------------|
| **Batch** | Batches telemetry data before sending | Improves compression and reduces network overhead | [Docs](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/batchprocessor) |
| **Memory Limiter** | Limits memory usage of the collector | Prevents out-of-memory conditions | [Docs](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/memorylimiterprocessor) |

### Resource & Attribute Processors

| Processor | Description | Use Case | Documentation |
|-----------|-------------|----------|---------------|
| **Resource** | Modifies resource attributes | Add, update, or delete resource-level attributes | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourceprocessor) |
| **Resource Detection** | Detects resource information from environment | Auto-detect cloud provider, k8s, and host information | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor) |
| **Attributes** | Modifies span, log, or metric attributes | Add, update, delete, or hash attribute values | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor) |
| **K8s Attributes** | Adds Kubernetes metadata to telemetry | Enrich data with pod, namespace, deployment info | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/k8sattributesprocessor) |
| **Group By Attrs** | Groups telemetry by specific attributes | Reorganize data streams based on attribute values | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/groupbyattrsprocessor) |

### Data Transformation Processors

| Processor | Description | Use Case | Documentation |
|-----------|-------------|----------|---------------|
| **Transform** | Transforms telemetry using OTTL | Complex transformations using OpenTelemetry Transformation Language | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor) |
| **Metrics Transform** | Renames and transforms metrics | Standardize metric names and units | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor) |
| **Cumulative To Delta** | Converts cumulative metrics to delta | Convert cumulative counters to rate-based metrics | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/cumulativetodeltaprocessor) |
| **Delta To Rate** | Converts delta metrics to rate | Calculate rates from delta values | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/deltatorate) |

### Filtering & Sampling Processors

| Processor | Description | Use Case | Documentation |
|-----------|-------------|----------|---------------|
| **Filter** | Filters telemetry based on conditions | Drop or include specific metrics, logs, or traces | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor) |
| **Probabilistic Sampler** | Samples traces based on probability | Reduce trace volume while maintaining statistical representation | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/probabilisticsamplerprocessor) |
| **Redaction** | Redacts sensitive information | Remove PII and sensitive data from telemetry | [Docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/redactionprocessor) |

---

## Exporters

Exporters send data to one or more backends or destinations.

| Exporter | Description | Use Case | Documentation |
|----------|-------------|----------|---------------|
| **OTLP** | Exports data via OTLP protocol (gRPC) | Send to OTLP-compatible backends (e.g., KloudMate) | [Docs](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter) |
| **OTLP HTTP** | Exports data via OTLP protocol (HTTP) | Send to OTLP HTTP endpoints | [Docs](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter) |
| **Debug** | Logs telemetry data to console | Development and troubleshooting | [Docs](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/debugexporter) |
| **Nop** | No-operation exporter that drops data | Testing and development scenarios | [Docs](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/nopexporter) |

---

## Connectors

Connectors connect two pipelines together, acting as both an exporter and a receiver.

Currently, no connectors are registered in the agent, but the framework supports them for future extensions.

---

For detailed configuration examples and parameters for each component, please refer to the [OpenTelemetry Collector Contrib documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib).

---
