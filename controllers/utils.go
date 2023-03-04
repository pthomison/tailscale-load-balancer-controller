package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ServiceReconciler) getService(ctx context.Context, name string, namespace string) (bool, *corev1.Service, error) {
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

func (r *ServiceReconciler) ensureConfigMap(ctx context.Context, cm *corev1.ConfigMap) error {

	name := types.NamespacedName{
		Name:      cm.Name,
		Namespace: cm.Namespace,
	}

	var tmp corev1.ConfigMap
	err := r.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, cm)
	} else {
		err = r.Update(ctx, cm)
	}
	return err
}

func (r *ServiceReconciler) ensureDeployment(ctx context.Context, d *appsv1.Deployment) error {

	name := types.NamespacedName{
		Name:      d.Name,
		Namespace: d.Namespace,
	}

	var tmp appsv1.Deployment
	err := r.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, d)
	} else {
		err = r.Update(ctx, d)
	}
	return err
}

func (r *ServiceReconciler) deleteConfigMap(ctx context.Context, namespacedName types.NamespacedName) error {
	var tmp corev1.ConfigMap
	err := r.Get(ctx, namespacedName, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		// Object Already Deleted
	} else {
		err = r.Delete(ctx, &tmp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ServiceReconciler) deleteDeployment(ctx context.Context, namespacedName types.NamespacedName) error {
	var tmp appsv1.Deployment
	err := r.Get(ctx, namespacedName, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		// Object Already Deleted
	} else {
		err = r.Delete(ctx, &tmp)
		if err != nil {
			return err
		}
	}

	return nil
}
