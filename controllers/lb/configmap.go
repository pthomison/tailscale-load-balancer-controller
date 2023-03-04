package lb

import (
	"bytes"
	"html/template"

	"github.com/pthomison/errcheck"
	"github.com/pthomison/tailscale-load-balancer-controller/controllers/names"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (LB *LoadBalancer) RenderConfigMap() {
	name, namespace, _ := names.TLBConfigMapName(LB.ServiceRequest)

	selectorLabelsMap, _ := names.SelectorLabels(LB.ServiceRequest.Name, LB.ServiceRequest.Namespace)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    selectorLabelsMap,
		},
		Data: map[string]string{
			"haproxy.cfg": renderHaproxyConfig(LB.Service),
		},
	}

	LB.ConfigMap = configMap
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
