package controllers

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/pthomison/errcheck"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func deploymentName(svc *corev1.Service) string {
	return fmt.Sprintf("tailscale-lb-%s", svc.Name)
}

func lbConfigMapName(svc *corev1.Service) string {
	return fmt.Sprintf("tailscale-lb-%s", svc.Name)
}

func lbKubeSecretName(svc *corev1.Service) string {
	return fmt.Sprintf("tailscale-lb-%s", svc.Name)
}

func SelectorLabels(svc *corev1.Service) (map[string]string, labels.Selector) {

	labelMap := make(map[string]string)

	common_key := "app.kubernetes.io/name"
	common_val := "tailscale-lb-provider"

	svc_key := "pthomison.com/lb-svc"
	svc_value := fmt.Sprintf("%s-%s", svc.Name, svc.Namespace)

	labelMap[common_key] = common_val
	labelMap[svc_key] = svc_value

	selector, err := labels.Parse(fmt.Sprintf("%s==%s,%s==%s", common_key, common_val, svc_key, svc_value))
	errcheck.Check(err)

	return labelMap, selector
}

func lbServiceAccountName() string {

	name := os.Getenv("SERVICE_ACCOUNT_NAME")

	if name == "" {
		return "testing-tailscale-pod"
	}

	return name
}

func tailscaleImage() string {

	image_tag := os.Getenv("TLB_IMAGE_TAG")

	if image_tag == "" {
		image_tag = "latest"
	}

	return fmt.Sprintf("pthomison/tailscale-lb:%s", image_tag)
	// return fmt.Sprintf("registry.localhost:15000/tailscale-lb:%s", image_tag)
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
