package controllers

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/pthomison/errcheck"
	corev1 "k8s.io/api/core/v1"
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

func lbServiceAccountName() string {

	name := os.Getenv("SERVICE_ACCOUNT_NAME")

	if name == "" {
		return "testing-tailscale-pod"
	}

	return name
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