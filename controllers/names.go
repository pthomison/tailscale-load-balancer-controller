package controllers

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/pthomison/errcheck"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func tlbNamespace() string {
	return "tailscale"
}

func tlbDeploymentName(req *ctrl.Request) (string, string, types.NamespacedName) {
	name := fmt.Sprintf("tlb-%s-%s", req.Namespace, req.Name)
	namespace := tlbNamespace()
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	return name, namespace, namespacedName
}

func tlbConfigMapName(req *ctrl.Request) (string, string, types.NamespacedName) {
	name := fmt.Sprintf("tlb-%s-%s", req.Namespace, req.Name)
	namespace := tlbNamespace()
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	return name, namespace, namespacedName
}

func tlbKubeSecretName(req *ctrl.Request) (string, string, types.NamespacedName) {
	name := fmt.Sprintf("tlb-%s-%s", req.Namespace, req.Name)
	namespace := tlbNamespace()
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	return name, namespace, namespacedName
}

func lbServiceAccountName() string {

	name := os.Getenv("SERVICE_ACCOUNT_NAME")

	if name == "" {
		return "testing-tailscale-pod"
	}

	return name
}

func SelectorLabels(svcName string, svcNamespace string) (map[string]string, labels.Selector) {

	labelMap := make(map[string]string)

	labelMap[commonLabel] = commonLabelVal
	labelMap[serviceNameLabel] = svcName
	labelMap[serviceNamespaceLabel] = svcNamespace

	selector, err := labels.Parse(fmt.Sprintf("%s==%s,%s==%s,%s==%s", commonLabel, commonLabelVal, serviceNameLabel, svcName, serviceNamespaceLabel, svcNamespace))
	errcheck.Check(err)

	return labelMap, selector
}

func tailscaleImage() string {

	image_tag := os.Getenv("TLB_IMAGE_TAG")
	if image_tag == "" {
		image_tag = "latest"
	}

	image_repo := os.Getenv("TLB_IMAGE_REPO")
	if image_repo == "" {
		image_repo = "registry.localhost:15000/tailscale-lb"
	}
	return fmt.Sprintf("%s:%s", image_repo, image_tag)
}

func renderHaproxyConfig(svc *corev1.Service) string {
	haproxy_template := `
# Global parameters
global
	maxconn 32000

	# Raise the ulimit for the maximum allowed number of open socket
	# descriptors per process. This is usually at least twice the
	# number of allowed connections (maxconn * 2 + nb_servers + 1) .
	ulimit-n 65535

	uid 0
	gid 0

	# daemon
	nosplice

# Default parameters
defaults
	# Default timeouts
	timeout connect 5000ms
	timeout client 50000ms
	timeout server 50000ms

{{ $url := printf "%s.%s.svc.cluster.local" .Name .Namespace }}

{{ range .Spec.Ports }}
frontend localhost-{{ .Port }}
	bind *:{{ .Port }}
	option tcplog
	mode tcp
	default_backend downstream-{{ .Port }}

backend downstream-{{ .Port }}
	mode tcp
	balance roundrobin
	server downstream {{ $url }}:{{ .Port }} check
{{ end }}
`

	tmpl, err := template.New("haproxy_template").Parse(haproxy_template)
	errcheck.Check(err)

	var buff bytes.Buffer
	tmpl.Execute(&buff, svc)

	return buff.String()
}
