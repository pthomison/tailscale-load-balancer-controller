#!/usr/bin/env bash

set -xe

echo "Debug"

echo "PodName: ${POD_NAME}"
echo "PodNamespace: ${POD_NAMESPACE}"

echo "Starting Tailscale Tunnel"

tailscaled --tun "userspace-networking" &
tailscale up --auth-key file:/opt/tailscale/token

echo "Started Tailscale Tunnel"

python3 \
	/root/update-ips.py \
		--namespace="${POD_NAMESPACE}" \
		--service="${SERVICE_NAME}" \
		--ip="$(tailscale ip -4)"