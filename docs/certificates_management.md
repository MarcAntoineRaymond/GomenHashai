# 🔐 Certificates Management

Certificates are required for the Validation and mutation Admission Webhooks and the metrics service.

By default a self signed CA and required certificates are generated with Helm.
But this is not a recommended setup.

You should use one of the following setup:

- With cert-manager and GomenHashai certificates resources

    GomenHashai Helm Chart comes with certificates resources to deploy with cert-manager. You can enable this mode by just setting cert-manager enabled to `true`.

    ```yaml
    certificates:
      cert-manager:
        enabled: true
    ```

    In addition you can customize duration, renewal and the issuer of these certificates.
    By default a cert-manager self-signed issuer will be created but you can set the issuer field to use an Issuer already created in your cluster.
    You can also configure the secretName for the certificates which is generated from the release Name by default.

    ```yaml
    certificates:
      # Cert manager configuration
      cert-manager:
        # Use cert-manager to inject CA in webhook configuration
        enabled: true
        # Deploy certificates resources for cert-manager (Only if cert-manager is enabled, disable creation if you provide your own certificates from a secret)
        create: true
        # Certificates duration in days
        duration: 365
        # When to renew certificates
        renewBefore: 360h
        # Only set issuer if you have your own issuer otherwise one will be created (if create is true)
        issuer:
      webhook:
        # Name of the secret containing webhook certificates (generated if empty)
        secretName:
      metrics:
        # Name of the secret containing metrics certificates (generated if empty)
        secretName:
    ```

- With cert-manager and your own certificates secrets

    With this option you provide your secrets name containing certificates managed by cert-manager for admission webhook and metrics endpoint.
    You need to enable cert-manager to configure CA injection in the admission webhook resources and disable GomenHashai certificates resources creation:

    ```yaml
    certificates:
      cert-manager:
        enabled: true
        create: false
      webhook:
        secretName: my-webhook-certs
      metrics:
        secretName: my-metrics-certs
    ```

- With your own certificates

    This option allows you to provide your own secrets containing your certificates without relying on a tool like cert-manager.
    However in this case you will need to to also provide the CA for both webhook configurations:

    ```yaml
    certificates:
      webhook:
        secretName: my-webhook-certs
      metrics:
        secretName: my-metrics-certs
    webhook:
      mutating:
        caBundle: |
          -----BEGIN CERTIFICATE-----
          ...
      validating:
        caBundle: |
          -----BEGIN CERTIFICATE-----
          ...
    ```

    Also you will need to handle certificate rotation/renewal manually, by updating the secret with your new certificates, updating the Helm Chart with the new CA bundle in the webhooks configurations and finnally restarting GomenHashai's pods.
