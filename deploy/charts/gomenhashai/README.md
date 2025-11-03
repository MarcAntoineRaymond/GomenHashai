# gomenhashai

![Version: 1.3.1](https://img.shields.io/badge/Version-1.3.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.3.1](https://img.shields.io/badge/AppVersion-v1.3.1-informational?style=flat-square)

Keep your Kubernetes cluster safe by ensuring all container's images use digests from a trusted set. GomenHashai verifies image integrity and gently apologizes as it gracefully denies or terminates pods that don’t meet the standard. Gomen Hashai~ 🙇

Built with security 🛡️ in mind, 🍣 GomenHashai ships with strong default protections.

## 🚀 Quick Start

Deploy in warn mode:

```sh
helm install gomenhashai gomenhashai --repo https://gomenhashai.github.io/GomenHashai \
  --namespace gomenhashai-system \
  --create-namespace \
  --set config.validationMode="warn"
```

---

## 🍣 Usage

GomenHashai uses **Kubernetes admission webhook** to validate and optionally mutate pod specifications to ensure all container images use **immutable digests** instead of mutable tags.
It helps enforce image provenance and strengthen your supply chain security posture.
It has many more features.

See the [Full Documentation](https://github.com/GomenHashai/GomenHashai).

**Homepage:** <https://github.com/GomenHashai/GomenHashai>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| MarcAntoineRaymond |  | <https://github.com/MarcAntoineRaymond> |

## Source Code

* <https://github.com/GomenHashai/GomenHashai>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| annotations | object | `{}` | Deployment annotations |
| args | list | `[]` | Args passed to the manager app, override default args |
| certificates.cert-manager | object | `{"create":true,"duration":365,"enabled":false,"issuer":null,"renewBefore":"360h"}` | Cert manager configuration |
| certificates.cert-manager.create | bool | `true` | Deploy certificates resources for cert-manager (Only if cert-manager is enabled, disable creation if you provide your own certificates from a secret) |
| certificates.cert-manager.duration | int | `365` | Certificates duration in days |
| certificates.cert-manager.enabled | bool | `false` | Use cert-manager to inject CA in webhook configuration |
| certificates.cert-manager.issuer | string | `nil` | Only set issuer if you have your own issuer otherwise one will be created (if create is true) |
| certificates.cert-manager.renewBefore | string | `"360h"` | When to renew certificates |
| certificates.duration | int | `365` | Certificates duration in days (for default self signed certificate) |
| certificates.metrics.secretName | string | `""` | Name of the secret containing metrics certificates (generated if empty) |
| certificates.webhook.secretName | string | `""` | Name of the secret containing webhook certificates (generated if empty) |
| config | object | `{}` | YAML configuration, see https://github.com/GomenHashai/GomenHashai?tab=readme-ov-file#-configurations |
| containerSecurityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"privileged":false,"readOnlyRootFilesystem":true,"runAsNonRoot":true}` | Container security context |
| digestsMapping | object | `{"create":true,"mapping":{},"secretKey":"digests_mapping.yaml","secretName":""}` | Mapping containing "image": "trusted digest" |
| digestsMapping.create | bool | `true` | Create the digestsMapping secret |
| digestsMapping.mapping | object | `{}` | YAML image name to digest mapping |
| digestsMapping.secretKey | string | `"digests_mapping.yaml"` | Name of the key under which the mapping is stored in the secret |
| digestsMapping.secretName | string | `""` | Name of the digestsMapping secret, if create is false secret must exist |
| envFrom | list | `[]` | Environment variables from secrets or configmaps to add to the container |
| extraEnv | list | `[]` | Extra environment variables to add to the container |
| extraLabels | object | `{}` | Extra labels |
| extraPodLabels | object | `{}` | Pod extra labels |
| extraVolumeMounts | list | `[]` | Extra volume mounts to add to the container |
| extraVolumes | list | `[]` | Extra volumes to add to the pod |
| fullnameOverride | string | `""` | Override ReleaseName-ChartName in template |
| globalPullSecrets | list | `[]` | Global image pull secrets to add to all namespaces |
| image.digest | string | `"sha256:75dec07996b338ba12d6986657a22ac560d7b3a4712fb53a4ab02f1aeca9b4d8"` | Image digest to use |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.repository | string | `"ghcr.io/gomenhashai/gomenhashai"` | Image repository |
| image.tag | string | `""` | Image tag to use, default to appVersion |
| imagePullSecrets | list | `[]` | Image pull secrets |
| initContainers | list | `[]` | Extra init containers to add to the pod |
| kubernetesClusterDomain | string | `"cluster.local"` | Cluster domain (used by cert-manager to generate certificate) |
| livenessProbe | object | `{"initialDelaySeconds":5,"periodSeconds":10,"port":8081}` | Configure Deployment liveness probe |
| metrics.enabled | bool | `true` | Enable exporting metrics with prometheus annotations |
| metrics.secure | bool | `true` | Serve metrics with HTTPS and authn/authz and add TLS Config to Service Monitor if enabled |
| metrics.service | object | `{"annotations":{},"extraLabels":{},"port":8443,"targetPort":8443,"type":"ClusterIP"}` | Metrics service configuration |
| metrics.serviceMonitor.annotations | object | `{}` | Custom annotations to add to ServiceMonitor |
| metrics.serviceMonitor.enabled | bool | `false` | Enable exporting metrics with Prometheus Service Monitor INSTEAD OF annotations, require using Prometheus Operator |
| metrics.serviceMonitor.endpointAdditionalProperties | object | `{}` | Extra properties to add to endpoint in ServiceMonitor resource |
| metrics.serviceMonitor.extraLabels | object | `{}` | Extra labels to add to ServiceMonitor |
| metrics.serviceMonitor.interval | string | `"60s"` | How frequently to scrape targets by default. |
| metrics.serviceMonitor.path | string | `"/metrics"` | Path exposing metrics |
| metrics.serviceMonitor.scrapeTimeout | string | `"10s"` | How long until a scrape request times out. |
| metrics.serviceMonitor.targetPort | int | `8443` | Port to scrape on metrics service |
| metrics.serviceMonitor.tlsConfig | object | `{}` | Overwrite Service Monitor TLS config, a default TLS config referencing certificates.metrcis.secretName is added when metrics.secure is true |
| nameOverride | string | `""` | Override Chart name in template |
| podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"gomenhashai"}` | Deployment pod annotations |
| podSecurityContext | object | `{"fsGroup":2000,"runAsGroup":2000,"runAsNonRoot":true,"runAsUser":1000,"seccompProfile":{"type":"RuntimeDefault"}}` | Pod security context |
| rbac | object | `{"create":true}` | RBAC role and binding to the service account |
| rbac.create | bool | `true` | Create the RBAC resources |
| readinessProbe | object | `{"initialDelaySeconds":5,"periodSeconds":10,"port":8081}` | Configure Deployment readiness probe |
| registriesConfig | object | `{}` | Registries authentication configuration, map of registry_name: {username: , password: } when automatically fetch digests is enabled |
| replicas | int | `1` | Replicas count multiple replicas is supported for HA |
| resources | object | `{"limits":{"cpu":"1","memory":"256Mi"},"requests":{"cpu":"10m","memory":"64Mi"}}` | Gomenhashai resources configuration, see https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits |
| serviceAccount | object | `{"annotations":{},"automountServiceAccountToken":true,"create":true,"extraLabels":{},"name":""}` | Service account configuration |
| serviceAccount.annotations | object | `{}` | Annotations to the service account if create is true |
| serviceAccount.automountServiceAccountToken | bool | `true` | Automount service account token in service account |
| serviceAccount.create | bool | `true` | Create the service account, if false the service account must be provided |
| serviceAccount.extraLabels | object | `{}` | Extra Labels to the service account if create is true |
| serviceAccount.name | string | `""` | Name of the service account, if create is false it must exists |
| sidecars | list | `[]` | Extra sidecars to add to the pod |
| tests.image.digest | string | `"sha256:0f6b5088710f1c6d2d41f5e19a15663b7fef07d89699247aaaad92975be7eed6"` |  |
| tests.image.repository | string | `"bitnami/kubectl"` |  |
| tests.image.tag | string | `"1.33.0-debian-12-r0"` |  |
| webhook.mutating | object | `{"annotations":{},"caBundle":"","enabled":true,"exemptNamespacesLabels":{},"extraLabels":{},"failurePolicy":"Fail","matchPolicy":"Exact","objectSelector":{},"reinvocationPolicy":"Never","sideEffects":"None"}` | Mutating Webhook configuration |
| webhook.mutating.caBundle | string | `""` | CA Bundle in PEM format to pass to the webhook, mandatory if not injected by cert-manager |
| webhook.mutating.enabled | bool | `true` | Enable mutation webhook |
| webhook.mutating.exemptNamespacesLabels | object | `{}` | Add labels: value to match namespace to exempt from mutation |
| webhook.service | object | `{"annotations":{},"extraLabels":{},"port":443,"targetPort":9443,"type":"ClusterIP"}` | Webhook service configuration |
| webhook.validating | object | `{"annotations":{},"caBundle":"","enabled":true,"exemptNamespacesLabels":{},"extraLabels":{},"failurePolicy":"Fail","matchPolicy":"Exact","objectSelector":{},"sideEffects":"None"}` | Validating Webhook configuration |
| webhook.validating.caBundle | string | `""` | CA Bundle in PEM format to pass to the webhook, mandatory if not injected by cert-manager |
| webhook.validating.enabled | bool | `true` | Enable validation webhook |
| webhook.validating.exemptNamespacesLabels | object | `{}` | Add labels: value to match namespace to exempt from validation |

