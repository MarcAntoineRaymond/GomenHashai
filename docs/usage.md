# 🍣 Usage

Main usage of GomenHashai is to enforce integrity in your Kubernetes Cluster using a list of trusted images digests.

But it can sastify many more use cases depending on how you configure it:

## Trusted Digests

You should carrefully pick your images and extract the digests from validated and secure images in your registry.

GomenHashai provides 2 ways to pass this digests list to the application:

- You let the helm chart create and mount the secret containing the digest list. You only needs to provide the mapping at the installation of the Helm Chart:

    *values.yaml*
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

    Using this option if you want to update the digest list you will need to redeploy the Helm Chart

- You create your own secret containing the mapping and only provides the secretName and Key to the Helm Chart installation:

    *secret*
    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: my-secret
    type: Opaque
    stringData:
      my-mapping.yaml: |
        "busybox:latest": "sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f"
        "busybox": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
        "library/busybox": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
        ...
    ```

    *values.yaml*
    ```yaml
    digestsMapping:
      # Create the secret
      create: false
      # Name of the secret, if create is false secret must exist
      secretName: my-secret
      # Name of the key under which the mapping is stored in the secret
      secretKey: my-mapping.yaml
   ```

   If you update the secret content you will need to restart GomenHashai pods to reload the new secret content.

### Digests Mapping content

Image name in the mapping that does not have a registry will match images from any registry. But if it contains a registry ex: `docker.io`, the image used in the pod should match the registry as well.

If the image in the mapping does not have a tag it will be used as default for this image if the container is using a tag that is not in the mapping. (This behaviour can be disabled, check [Configurations](../README.md#-configurations))

For instance with the following mapping:

```yaml
"library/busybox:1": "sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f"
"docker.io/library/nginx": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
```

If you run a container using image `library/busybox:1` it will be allowed and the digest `sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f` will be added to ensure the right image is used.

Now if you run a container with image `library/busybox:2` it will be denied.

Since there are no registry defined, running containers with these images with any registry will have the same results:

- `docker.io/library/busybox:1` allowed and digest is added
- `docker.io/library/busybox:2` denied

Now for image `docker.io/library/nginx` we specified the registry in the mapping so we get the following results:

- `library/nginx` denied
- `docker.io/library/nginx` allowed and digest is added
- `docker.io/library/nginx:1` allowed and digest is added (as there were no tag defined in the mapping)
- `library/nginx:1` denied

Be careful with the tags and registry, very often the same image will have different digests in different registry and tags cannot be easily swapped.
In most cases you may want to specify both tags and registry in mapping.

## Fetch digests from registry

Instead of using a secret listing trusted digests, you can automatically fetch digests from your image registry:

```yaml
config:
  fetchDigests: true
```

*Note: This mode is less secure because image digests are not pre-verified.
Additionally, the webhook must contact the registry to retrieve the digest, which may slow down pod validation depending on network latency and registry response times.*

This mode is best suited for environments where all images originate from a secure, trusted, and verified registry.

GomeHashai will fetch digests based on the registry specified in the image reference. You can enforce a specific registry using the Registry Mutation feature.

### Exporting Digests for Trusted Use

With the digests automatically fetched from the registry you could use a bash command to extract the digests and images to make a mapping usable with the trusted digest secret:

```sh
kubectl get pods --all-namespaces -o json | jq -r '
  .items[]
  | (
      .spec.containers[]?,
      .spec.initContainers[]?
    )
  | select(.image | test("@sha256:"))
  | "\"\(.image | split("@")[0])\": \"\(.image | split("@")[1])\""
' | sort | uniq
```

*This command require kubectl and jq.*

Sample output:

```sh
"docker.io/grafana/grafana:11.6.1": "sha256:52c3e20686b860c6dc1f623811565773cf51eefa378817a4896dfc863c3c82c8"
"registry.k8s.io/etcd:3.5.16-0": "sha256:c6a9d11cc5c04b114ccdef39a9265eeef818e3d02f5359be035ae784097fdec5"
"registry.k8s.io/kube-apiserver:v1.32.0": "sha256:ebc0ce2d7e647dd97980ec338ad81496c111741ab4ad05e7c5d37539aaf7dc3b"
```

You can use this output to populate the trusted digest secret. Once you've validated each digest, you may disable automatic fetching to enforce stronger security.

## Audit or Dry Run

Enforcing behaviour of the mutating and validating webhook can be disabled.

This is very useful if you do not want to delete or deny any pods in the cluster.
This could be the case if you want to check if your environment is compliant before potentially breaking anything if it is not.

To disable enforcing mode for the validation and not deny or delete pods you can set the following variable in your `config`:

```yaml
# -- Can be warn or fail (default)
validationMode: "warn"
```

You will get warning when creating pods that are not using trusted digests and GomenHashai will log the event.

The following variable in your `config` will stop GomenHashai from appending digests from your trusted secret to pods container images, but it will still logs the event:

```yaml
# -- Enable to not modify pods but instead logs (pods will fail validation unless you disable it or set it in warn)
mutationDryRun: true
```

You can also completely disable both webhooks from the Helm Chart values but in this case pods will not be submitted to any check and GomenHashai will not be able to log anything.

### Exemptions

It is possible to exempt a list of images, or even use regex to exempt images by setting the variable `exemptions` in the Helm Chart config:

```yaml
config:
  exemptions:
    - ".*redis:.*"
    - "docker.io/library/busybox:12"
```

The Helm Chart will exempt the namespace in which you install 🍣GomenHashai, you can exempt other namespaces as well:

```yaml
webhook:
  mutating:
    exemptNamespacesLabels:
      kubernetes.io/metadata.name:
        - "kube-system"
        - "cert-manager"
  validating:
    exemptNamespacesLabels:
      kubernetes.io/metadata.name:
        - "kube-system"
        - "cert-manager"
```

## 📈 Monitoring

GomenHashai exposes useful metrics. You could know how many pods were denied or allowed for instance. Refer to the [monitoring section](monitoring.md)
