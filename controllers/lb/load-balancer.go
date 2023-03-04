package lb

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type LoadBalancer struct {
	ServiceRequest *ctrl.Request
	Service        *corev1.Service

	Deployment *appsv1.Deployment
	ConfigMap  *corev1.ConfigMap
}

func (LB *LoadBalancer) Render() {
	LB.RenderDeployment()
	LB.RenderConfigMap()
}
