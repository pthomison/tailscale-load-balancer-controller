#  tailscale-load-balancer-controller
A k8s load balancer service controller which allows you to expose kubernetes services to a tailscale network

## Description
For every service of type `LoadBalancer`, the controller will launch a tailscale connected pod && forward traffic from that pod to the service. Additionally, the controller will populate the `externalIPs` field of the service with the assigned tailscale IP.

## Getting Started

Requirements (hard codes that need to be broken out at somepoint):
- a namespace with the name `tailscale`
- a secret in that namespace with the name `tailscale-token` with a key of `token` that holds an tailscale auth key (preferably ephemeral)

Deployment Options:
- `make deploy` will use kustomize to template the artifacts into your cluster
- `make template > your-spot-for-config.yaml` will just template out the artifacts, allowing you to place them whereever fits into your deployment pipeline
- helm chart, see below

### Helm

Currently there is a helm chart, but its relatively unconfigurable as its just `make template` stored under `templates/raw.yaml`. In the future, having a full fledged helm chart could definetly be worth while, so as needs arise, sections may be broken out of `raw.yaml` into a normal, configurable helm template.

To Use:

```sh
helm repo add tailscale-load-balancer-controller https://pthomison.github.io/tailscale-load-balancer-controller
helm repo update
helm install tailscale-load-balancer-controller tailscale-load-balancer-controller/tailscale-load-balancer-controller
```

## ToDo

- ~~stop using "latest" for the deployed LB pod~~
- better helm chart
- ~~GH actions work, make sure image & chart publishing is working~~
- Better "ip-updater" solution/loop
- Configurable userspace vs kernel wireguard
- Stop having trash commit messages on main
- Scope down service account permissions
- ~~Separate service account for lb pods~~
- ~~Garbage Collection of lb pods~~
- ~~Easy way to toggle the use of dev images vs dockerhub~~
- Testing


## How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

