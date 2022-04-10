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
parser.add_argument("--service", required=True, help="Service Name To Inject IP To")
parser.add_argument("--namespace", required=True, help="Service Namespace")

args = parser.parse_args()
IP = args.ip
SERVICENAME = args.service
NAMESPACE = args.namespace

external_ip = {
    "spec": {
        "externalIPs": [IP],
    }
}

### K8S Setup
config.load_incluster_config()
coreV1 = client.CoreV1Api()

### Main Loop
logging.info(f"Starting Update IP Loop")
logging.info(f"IP To Inject: {IP}")

while True:
    # find self-service
    service = coreV1.read_namespaced_service(SERVICENAME, NAMESPACE)
    # logging.info(f"Service: {service.metadata.name}/{service.metadata.namespace} : { ', '.join(service.spec.external_i_ps) }")

    if service.spec.external_i_ps == None or service.spec.external_i_ps[0] != IP:
        patchedService = coreV1.patch_namespaced_service(
            SERVICENAME, NAMESPACE, external_ip
        )
        logging.info(
            f"Patched Service: {patchedService.metadata.name}/{patchedService.metadata.namespace} : {', '.join(service.spec.external_i_ps) }"
        )

    time.sleep(30)
