# cartographer
An application to manage upstream, third party kubernetes manifests via a remote kustomization.yaml file or a remote helm chart. Cartographer will notify you of upstream changes and help you merge them into your infrastructure as code repositories.

## Use Cases
Using this readme as a living document to help with project planning.

### Upstream Helm Chart
1. Check for updates to the helm chart
2. convert to kustomize and apply transformations and rule sets. This could be things like removing annotations and labels or ignoring clusterrolebindings
3. Make a PR against the kustomize base

### Upstream Kustomization.yaml
1. Check for new tags
2. make a PR to update the tag

### Raw Kubernetes Manifests
1. Check for updates
2. apply transformations and rule sets
3. Make a PR against the kustomize base

## Design
- hold state in CRDs
- microservices with redis pod
- designed to run on k8s
- add frontend in v2, if we take the PR bot approach this could be pretty autonomous

## TODO
- design data structure
