# tailscale ingress controller

A fairly simplistic wrapper around the nginx ingress controller that adds a connection into a tailscale network && injects the tailscale IP into the ingress controller service

### Quickstart

Add the helm repo
```
helm repo add tailscale-ingress-controller https://pthomison.github.io/tailscale-ingress-controller
helm repo update
```

Inject your token
```
tbd
```

Install the controller
```
helm install tailscale-ingress-controller tailscale-ingress-controller/tailscale-ingress-controller
```