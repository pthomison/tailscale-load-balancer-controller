package controllers

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/pthomison/errcheck"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func lbPodName(svc *corev1.Service) string {
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

func NewLB(svc *corev1.Service) (*corev1.Pod, *corev1.ConfigMap) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lbPodName(svc),
			Namespace: defaultNamespace,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: lbServiceAccountName(),
			Containers: []corev1.Container{
				{
					Name:  "tailscale",
					Image: "tailscale/tailscale:stable",
					Env: []corev1.EnvVar{
						{
							Name: "TS_AUTHKEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: defaultSecret,
									},
									Key: defaultSecretKey,
								},
							},
						},
						{
							Name:  "TS_KUBE_SECRET",
							Value: lbKubeSecretName(svc),
						},
						{
							Name:  "TS_ACCEPT_DNS",
							Value: "false",
						},
						// {
						// 	Name: "POD_NAME",
						// 	ValueFrom: &corev1.EnvVarSource{
						// 		FieldRef: &corev1.ObjectFieldSelector{
						// 			FieldPath: "metadata.name",
						// 		},
						// 	},
						// },
						// {
						// 	Name: "POD_NAMESPACE",
						// 	ValueFrom: &corev1.EnvVarSource{
						// 		FieldRef: &corev1.ObjectFieldSelector{
						// 			FieldPath: "metadata.namespace",
						// 		},
						// 	},
						// },
					},
				},
				{
					Name:  "haproxy",
					Image: "haproxy:2.7",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "haproxy-config",
						MountPath: "/usr/local/etc/haproxy",
						ReadOnly:  true,
					}},
				},
			},
			Volumes: []corev1.Volume{{
				Name: "haproxy-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: lbConfigMapName(svc),
						},
					},
				},
			}},
		},
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lbConfigMapName(svc),
			Namespace: defaultNamespace,
		},
		Data: map[string]string{
			"haproxy.cfg": renderHaproxyConfig(svc),
		},
	}

	return pod, configMap
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
