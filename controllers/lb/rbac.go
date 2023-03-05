package lb

import (
	"github.com/pthomison/tailscale-load-balancer-controller/controllers/names"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (LB *LoadBalancer) RenderServiceAccount() {
	name, namespace, _ := names.TLBNamespacedName(LB.ServiceRequest)

	selectorLabelsMap, _ := names.SelectorLabels(LB.ServiceRequest.Name, LB.ServiceRequest.Namespace)

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    selectorLabelsMap,
		},
	}

	LB.ServiceAccount = sa
}

func (LB *LoadBalancer) RenderRole() {
	name, namespace, _ := names.TLBNamespacedName(LB.ServiceRequest)

	selectorLabelsMap, _ := names.SelectorLabels(LB.ServiceRequest.Name, LB.ServiceRequest.Namespace)

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    selectorLabelsMap,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "update"},
			},
		},
	}

	LB.Role = role
}

func (LB *LoadBalancer) RenderRoleBinding() {
	name, namespace, _ := names.TLBNamespacedName(LB.ServiceRequest)

	selectorLabelsMap, _ := names.SelectorLabels(LB.ServiceRequest.Name, LB.ServiceRequest.Namespace)

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    selectorLabelsMap,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     name,
		},
	}

	LB.RoleBinding = rb
}
