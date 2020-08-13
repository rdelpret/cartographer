# Cartographer
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

#### MVP Spec
```sources:
- name: aws-auto-discover-manifest
  type: manifest
  recurse: false
  github: kubernetes/autoscaler
  path: cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml
  branch: master

destinations:
-  name: infra
   github: rdelpret/cartographer-infra-test-repo
   path: kustomize/cluster-autoscaler/base
   type: kustomize
-  name: namespaces
   github: rdelpret/cartographer-infra-test-repo
   path: namespaces
   type: manfiest

route:
- source: '*'
  objectTypes:
  - namespace
  destination: namespaces
  transformations:
  - cartograher-label
  method: PR
- source: '*'
  objectTypes: '*'
  destination: infra
  transformations: []
  method: PR

transformations:
- name: cargographer-label
  addLabel: cartographer
```

## Design
- hold state in CRDs
- microservices with redis pod
- designed to run on k8s
- add frontend in v2, if we take the PR bot approach this could be pretty autonomous

## TODO
- [x] design basic data structure 
- [x] write code to unmarshall yaml into a struct
- [x] create test infra repo cartographer-infra-test-repo
- [x] initalize empty directories in test repo
- [x] loop through multiple yaml files and create a list of structs
- [x] impliment multiple sources
- [x] impliment multiple destinations
- [x] impliment mutliple routes
- [ ] pull github token in from secret, file to start
- [ ] figure out what contexts do
- [ ] create an abstraction for the git client so I can use it in multiple places, or just a flow for writing git functions
- [ ] Reach out to github and pull the latest master of that file, write the commit to the app
- [ ] make an inital PR against the test repo
