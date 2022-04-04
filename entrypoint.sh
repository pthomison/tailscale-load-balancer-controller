#!/usr/bin/env bash

set -xe

echo "Starting Tailscale Tunnel"

tailscaled --tun "userspace-networking" &
tailscale up --auth-key file:/opt/tailscale/token

echo "Started Tailscale Tunnel"

python3 \
	/root/update-ips.py \
		--namespace="tailscale" \
		--service="tailscale-ingress-controller-ingress-nginx-controller" \
		--ip="$(tailscale ip -4)"