#!/usr/bin/env bash

set -xe

echo "PodNamespace: ${POD_NAMESPACE}"

# Wait for tailscale tunnel
echo "Waiting For Tailscale Tunnel"

until ifconfig | grep -A 1 "tailscale0" | egrep -o "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b" | uniq | grep "100.";
do
	sleep 1
done

TAILSCALE_IP="$(ifconfig | grep -A 1 "tailscale0" | egrep -o "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b" | uniq | grep "100.")"

echo "Tunnel Established; IP=${TAILSCALE_IP}"

echo "Starting IP Updater"
python3 \
	/root/update-ips.py \
		--namespace="${POD_NAMESPACE}" \
		--service="${SERVICE_NAME}" \
		--ip="${TAILSCALE_IP}"
