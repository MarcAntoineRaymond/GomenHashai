# 📈 Monitoring

GomenHashai exposes useful Prometheus-compatible metrics. Configuration options including port, security, and authorization are available via the Helm chart.

## Enabling / Disabling Metrics
Metrics collection is enabled by default. To disable it, set the following in your Helm values:

```yaml
metrics:
  enabled: false
```

## HTTPS and Authorization

By default, the metrics endpoint is secured with HTTPS and requires an authenticated Kubernetes `ServiceAccount` token with appropriate RBAC permissions.

To expose metrics over plain HTTP without authentication, use:

```yaml
metrics:
  secure: false
```

> ⚠️ Disabling security exposes the metrics endpoint without authentication or encryption. Use only in trusted environments

## Prometheus Service Monitor

If you're using the Prometheus Operator, you can enable a ServiceMonitor by setting:

```yaml
metrics:
  serviceMonitor:
    enabled: true
```

> ❗Requires Prometheus Operator to be installed in the cluster.

When `secure` is enabled, the `ServiceMonitor` is automatically configured with TLS and a bearer token to access the endpoint.

### TLS Configuration

The Certificate Authority (CA) used to verify the endpoint's TLS certificate is taken from the secret specified in:

```yaml
certificates:
  metrics:
    secretName:
```

Take a look a the [certificate management section](certificates_management.md) for details on the cerificates secrets configuration.

The secret must contain the CA certificate under the key ca.crt.

To customize the TLS configuration for the ServiceMonitor, you can override the `tlsConfig` field:

```yaml
metrics:
  serviceMonitor:
    tlsConfig: {}
```

Refer to the [Prometheus documentation](https://prometheus-operator.dev/docs/api-reference/api/#monitoring.coreos.com/v1.TLSConfig) for valid tlsConfig fields.

## Exposed Metrics

GomenHashai exposes common metrics, `controller-runtime` metrics (e.g., reconciliation, queue length, etc.) and some custom metrics:

|Metric Name|Description|
|----|-----|
| gomenhashai_validation_total | Number of pods processed by GomenHashai's validating webhook |
|gomenhashai_mutation_total|Number of pods processed by GomenHashai's mutation webhook|
|gomenhashai_allowed_count|Number of pods Allowed by GomenHashai|
|gomenhashai_denied_count|Number of pods Denied by GomenHashai|
|gomenhashai_warnings_count|Number of pods processed with Warnings by GomenHashai|
|gomenhashai_mutation_exempted_count|Number of pods Exempted processed by GomenHashai during mutation|
|gomenhashai_validation_exempted_count|Number of pods Exempted processed by GomenHashai during validation|
|gomenhashai_deleted_count|Number of pods Deleted by GomenHashai|
