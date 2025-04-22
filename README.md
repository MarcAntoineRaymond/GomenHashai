# 🍣 GomenHashai 🐾

![GomenHashai Logo](logo/logo.png)

Keep your Kubernetes cluster safe by ensuring all pod images use digests from a trusted set. GomenHashai verifies image integrity and gently apologizes as it gracefully denies or terminates pods that don’t meet the standard. Gomen Hashai~ 🙇

Built with security 🛡️ in mind, 🍣 GomenHashai ships with strong default protections.


*✨ “GomenHashai” is a playful mix of “Gomen nasai” (ごめんなさい – Japanese for “I’m sorry”) and “Hash,” referencing image digests. Because it says sorry when it denies your pods 😄. 🍣*

---

## 📚 Table of Contents

- [✨ What It Does](#-what-it-does)
- [🔧 Configurations](#-configurations)
- [🚀 Deployment](#-deployment)
- [📄 License](#-license)

---

## ✨ What It Does

### 🌀 Mutating Admission Webhook

Automatically rewrites container images in Pods to include a trusted digest (e.g., nginx:latest -> nginx:latest@sha256:...).

### 🛡️ Validating Admission Webhook

Denies Pods that contain containers without trusted digests.

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

```

Image name in the mapping that does not have a registry will match images from any registry.
But if it contains a registry ex: `docker.io` the image used in the pod should match the registry as well.

If the image in the mapping does not have a tag it will be used as default for this image if the container is using a tag that is not in the mapping. (This behaviour can be disabled)

### 📦 Helm Chart

Deploy the entire setup in one command with Helm.

Includes webhook deployment, certificates (with cert-manager), and RBAC.

The provided Helm chart follows Kubernetes security best practices out of the box.

### 📰 Logging

Follow exactly what resources gets denied, deleted or modified in the logs:

Messages using `[🍣GomenHashai!]` and `❌` indicates a pod was denied and message `[🍣GomenHashai] integrity verified` indicates the pod will be authorized.

Messages using `[🐾IntegrityPatrol]` are informative.

### 🐳 Registry Modification

Mutating webhook can also be used to enforce a common registry for all images.

### ⛩️ Exemptions

It is possible to exempt a list of images, or even use regex to exempt images.

The Helm Chart will exempt the namespace in which you install 🍣GomenHashai, you can exempt other namespaces as well.

### ⚠️ Warning

Running without configuring much may break stuff in your cluster, 🍣GomenHashai 🙇. Be careful providing a complete mapping for images/digests or set exemptions for your important cluster services (CNI, control-plane, api server...).

---

## 🔧 Configurations

A YAML configuration file can be used to customize the processing behaviour in addition to the Helm Chart configuration:

```yaml
# -- Path to the digests mapping file
digestsMappingFile: "/etc/gomenhashai/digests/digests_mapping.yaml"
# -- List of images to skip, can contain regex ex: ".*redis:.*"
exemptions: []
# -- If the image in the mapping does not have a tag it will be used as default for this image if the container is using a tag that is not in the mapping
imageDefaultDigest: true
# -- Can be warn or fail (default)
validationMode: "fail"
# -- Enable to not modify pods but instead logs (pods will fail validation unless you disable it or set it in warn)
mutationDryRun: false
# -- Enable modifying the registry part of images with the value of MutationRegistry
mutationRegistryEnabled: false
# -- The registry to inject when MutationRegistryEnabled is true
mutationRegistry: ""
# -- Configuration of the process that handles existing pods on init
existingPods:
# -- Enable the init function that will process existing pods at startup
    enabled: true
# -- Timeout used to wait before starting this job in seconds
    startTimeout: 5
# -- Timeout used to wait before retrying to process pods that failed in seconds
    retryTimeout: 5
# -- How many times we should retry processing pods that failed
    retries: 5
# -- Replace already existing pods with output from webhook, if disabled webhook will be used with dry run to not modify pods
    updateEnabled: true
# -- Allow deleting existing pods that are forbidden by webhook
    deleteEnabled: true
```

The configuration file path can be overwritten by the environment variable `GOMENHASHAI_CONFIG_PATH`.

Using this configuration it is possible to disable the job that process existing pods: `existingPods.enabled`

It is also possible to run this tool without blocking pods: `validationMode: warn`

Each variable can be overwritten by an environment variable.

The variable starts with `GOMENHASHAI_` and follows with the variable name in upper case: `GOMENHASHAI_VALIDATIONMODE` or `GOMENHASHAI_EXISTING_PODS_ENABLED`, ommitting the `GOMENHASHAI_` will also work but it is better to keep it.

---

## 🚀 Deployment

### 🚀 Deploy with Helm

[Deploy cert-manager in your cluster](https://cert-manager.io/docs/installation/) (You can also provide your own certificates).

Add the GomenHashai repo and install the Chart:

```sh
helm repo add gomenhashai https://marcantoineRaymond.github.io/GomenHashai
helm install my-release gomenhashai/gomenhashai --set "certificates.cert-manager.enabled=true"
```

You need to provide the digest mapping in the values:

```yaml
digestsMapping:
  mapping:
    "busybox:latest": "sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f"
    "busybox": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
    "library/busybox": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
    "docker.io/library/busybox": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
    "docker.io/library/busybox:stable": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
    "busybox:stable": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
    "nginx/nginx-ingress:5.0.0-alpine": "sha256:a6c4d7c7270f03a3abb1ff38973f5db98d8660832364561990c4d0ef8b1477af"
    "curlimages/curl:8.13.0": "sha256:d43bdb28bae0be0998f3be83199bfb2b81e0a30b034b6d7586ce7e05de34c3fd"
```

You could also provide your trusted digests mapping from an already created secret, it needs to be created in the same namespace you deploy:

```yaml
digestsMapping:
  # Create the secret
  create: false
  # Name of the secret, if create is false secret must exist
  secretName: my-secret
```

### 🛠️ Build Locally

Clone the repo:

```sh
git clone https://github.com/yourusername/gomenhashai.git
cd gomenhashai
```

Build the binary:

```sh
make docker-build docker-push IMG=<your_image>
```

Install the chart:

```sh
helm install gomenhashai ./deploy/charts/gomenhashai
```

## ⚙️ Helm Chart Values


---

## 📄 License

This project is licensed under the **Apache License 2.0**. See the [LICENSE](./LICENSE) and [NOTICE](./NOTICE) files for details.
