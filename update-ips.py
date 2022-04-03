#!/usr/bin/env python3

import logging
import time
import argparse
from kubernetes import client, config, watch
import pprint

logging.basicConfig(level=logging.INFO)

### CLI Flags

parser = argparse.ArgumentParser()

parser.add_argument("--ip", required=True, help="IP To Update Ingresses With")
parser.add_argument("--ingressclass", required=True, help="Ingress Class To Target")

args = parser.parse_args()

NAMESPACE="tailscale"
SERVICENAME="tailscale-ingress-controller-ingress-nginx-controller"

IP=args.ip
INGRESSCLASS=args.ingressclass

external_ip = {
	"spec": {
		"externalIPs": [IP]
	}
}

### K8S Setup
config.load_incluster_config()
networkingV1 = client.NetworkingV1Api()
coreV1 = client.CoreV1Api()

### Main Loop
logging.info(f"Starting Update IP Loop")
logging.info(f"Ingress Class Target: {INGRESSCLASS}")
logging.info(f"IP To Inject: {IP}")

while True:
	# find self-service
	service = coreV1.read_namespaced_service(SERVICENAME, NAMESPACE)

	if service.spec.external_i_ps != IP:
		patchedService = coreV1.patch_namespaced_service(SERVICENAME, NAMESPACE, external_ip)
		pprint.pprint(patchedService)

	time.sleep(30)
