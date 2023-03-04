package lb

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pthomison/tailscale-load-balancer-controller/controllers/names"
)

func (LB *LoadBalancer) RenderDeployment() {

	cfName, _, _ := names.TLBNamespacedName(LB.ServiceRequest)
	tsKubeSecretName, _, _ := names.TLBNamespacedName(LB.ServiceRequest)

	deploymentName, deploymentNamespace, _ := names.TLBNamespacedName(LB.ServiceRequest)

	serviceAccountName, _, _ := names.TLBNamespacedName(LB.ServiceRequest)

	selectorLabelsMap, _ := names.SelectorLabels(LB.ServiceRequest.Name, LB.ServiceRequest.Namespace)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: deploymentNamespace,
			Labels:    selectorLabelsMap,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabelsMap,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selectorLabelsMap,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{
						{
							Name:            "tailscale",
							Image:           names.TLBImage(),
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name: "TS_AUTHKEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: names.DefaultSecret,
											},
											Key: names.DefaultSecretKey,
										},
									},
								},
								{
									Name:  "TS_KUBE_SECRET",
									Value: tsKubeSecretName,
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
							Name:            "ip-reflector",
							Image:           names.TLBImage(),
							ImagePullPolicy: corev1.PullAlways,
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
										Name: cfName,
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

	LB.Deployment = deployment
}
