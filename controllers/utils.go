package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ServiceReconciler) GetService(ctx context.Context, name string, namespace string) (bool, *corev1.Service, error) {
	var svc corev1.Service
	err := r.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, &svc)

	if client.IgnoreNotFound(err) != nil {
		return false, nil, err
	} else if err != nil {
		return false, nil, nil
	} else {
		return true, &svc, nil
	}
}
