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
SERVICENAME="nginx-ingress-ingress-nginx-controller"

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

# find self-service

service = coreV1.read_namespaced_service(SERVICENAME, NAMESPACE)
pprint.pprint(service)

pprint.pprint(external_ip)

patchedService = coreV1.patch_namespaced_service(SERVICENAME, NAMESPACE, external_ip)
pprint.pprint(patchedService)


# watch ingresses for changes
# w = watch.Watch()
# for event in w.stream(networkingV1.list_ingress_for_all_namespaces):
# 	logging.info("noop")

# 	ingress_name = event['object'].metadata.namespace
	
# 	ingress_namespace = event['object'].metadata.namespace
# 	ingress_class = event['object'].spec.ingress_class_name

# 	logging.info(f"Ingress Name: {ingress_name}")
# 	logging.info(f"Ingress Namespace: {ingress_namespace}")
# 	logging.info(f"Ingress Class: {ingress_class}")
