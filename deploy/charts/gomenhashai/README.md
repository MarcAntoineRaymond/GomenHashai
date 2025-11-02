# 🍣 GomenHashai 🐾

Keep your Kubernetes cluster safe by ensuring all container's images use digests from a trusted set. GomenHashai verifies image integrity and gently apologizes as it gracefully denies or terminates pods that don’t meet the standard. Gomen Hashai~ 🙇

Built with security 🛡️ in mind, 🍣 GomenHashai ships with strong default protections.

## 🚀 Quick Start

```sh
helm install gomenhashai gomenhashai --repo https://marcantoineRaymond.github.io/GomenHashai \
  --namespace gomenhashai-system \
  --create-namespace \
  --set config.validationMode="warn"
```

---

## ✨ What It Does

### 🌀 Mutating Admission Webhook

Automatically rewrites container images in Pods to include a trusted digest (e.g., nginx:1.27.5 -> nginx:1.27.5@sha256:...).

### 🛡️ Validating Admission Webhook

Denies Pods that uses containers without trusted digests.

Ensures every container image matches a digest listed in a trusted Secret.

### ↩️ Handles Already Existing pods

Can submit automatically already existing pods to the webhook to make sure they use a digest. It can potentially delete pods using untrusted digests/images.

### 🔐 Trusted Digest Store

Uses a Kubernetes Secret to store a mapping of image name -> digest.

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gomenhashai-digests-mapping
type: Opaque
stringData:
  digests_mapping.yaml: |
    "busybox:latest": "sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f"
    "busybox": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
    "library/busybox": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
    ...
```

### 🔃 Fetch digests from registry

Instead of using a secret listing trusted digests, you can automatically fetch digests from your image registry:

```yaml
config:
  fetchDigests: true
```

### 📈 Monitoring

GomenHashai exposes useful custom Prometheus-compatible metrics.
You could get metrics helping understand how many pods are compliant with digests.
Configuration options including port, security, and authorization are available via the Helm chart.

### 📦 Helm Chart

Deploy the entire setup in one command with Helm.

Includes webhook deployment, certificates (with cert-manager), and RBAC.

The provided Helm chart follows Kubernetes security best practices out of the box.

### 🐳 Registry Modification

Mutating webhook can also be used to enforce a common registry for all images.
In addition to the registry, the pullPolicy and imagePullSecrets can also be enforced for all pods.

### ⛩️ Exemptions

It is possible to exempt a list of images, or even use regex to exempt images.

The Helm Chart will exempt the namespace in which you install 🍣GomenHashai, you can exempt other namespaces as well.
