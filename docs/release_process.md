# 📦 Release Process

GomenHashai release process involves building and publishing a Docker image, along with releasing an updated Helm chart for deployment. Key aspects of the release workflow are as follows:

## 🐳 Docker Image

- Every release of the application results in a new Docker image.
- Images are tagged using Semantic Versioning (SemVer), e.g., v1.2.3.
- These image tags reflect the application version and are used for traceability and deployment configuration.
- Additionally, a main tag is pushed whenever changes are merged into the main branch.
⚠️ Note: The main tag is not a stable release and may be broken or unstable. It is primarily intended for development and testing purposes.
- Docker images are published to: `ghcr.io/marcantoineraymond/gomenhashai`

## ☸️ Helm Chart

- A Helm chart is maintained for deployment of the application to Kubernetes clusters.
- The Helm chart has its own independent versioning, also following SemVer but without the v_ prefix.
- A new chart version is released every time a new Docker image is built, even if there are no Helm-specific changes. This ensures that the chart can reference the latest image with it's specific digest.
- In some cases, the chart is updated to introduce new configuration options, templates, or other Helm-related improvements. These chart-only updates may not trigger a new image release.
- Helm charts are published to the following Helm chart repository: `https://marcantoineRaymond.github.io/GomenHashai`

## Version Alignment

- Docker image versions and Helm chart versions are not tightly coupled.
- While each Helm release generally references a specific image version, Helm chart updates can occur independently of application changes.

This separation allows for flexible iteration on deployment strategies and configuration while maintaining clear tracking of application version changes.
