package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tailscaleImage = "registry.localhost:15000/tailscale-lb:7187c3e"
)

func (lb *LoadBalancer) renderDeployment() {
	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName(lb.svc),
			Namespace: defaultNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": "tailscale-lb-provider",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": "tailscale-lb-provider",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: lbServiceAccountName(),
					Containers: []corev1.Container{
						{
							Name:  "tailscale",
							Image: tailscaleImage,
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
									Value: lbKubeSecretName(lb.svc),
								},
								{
									Name:  "TS_ACCEPT_DNS",
									Value: "false",
								},
							},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "tailscale-socket",
								MountPath: "/tmp",
							}},
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
						{
							Name:  "ip-reflector",
							Image: tailscaleImage,
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "tailscale-socket",
								MountPath: "/tmp",
							}},
							Command: []string{
								"/ip-monitor-entrypoint.sh",
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "haproxy-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: lbConfigMapName(lb.svc),
									},
								},
							},
						},
						{
							Name: "tailscale-socket",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	lb.LoadBalancerObjects.Deployment = &deployment
}

func (lb *LoadBalancer) renderConfigMap() {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lbConfigMapName(lb.svc),
			Namespace: defaultNamespace,
		},
		Data: map[string]string{
			"haproxy.cfg": renderHaproxyConfig(lb.svc),
		},
	}

	lb.LoadBalancerObjects.ConfigMap = configMap
}
