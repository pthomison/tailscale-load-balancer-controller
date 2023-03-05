package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

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
	err := r.UncachedClient.Get(ctx, name, &tmp)
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
	err := r.UncachedClient.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, d)
	} else {
		err = r.Update(ctx, d)
	}
	return err
}

func (r *ServiceReconciler) ensureServiceAccount(ctx context.Context, sa *corev1.ServiceAccount) error {

	name := types.NamespacedName{
		Name:      sa.Name,
		Namespace: sa.Namespace,
	}

	var tmp corev1.ServiceAccount
	err := r.UncachedClient.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, sa)
	} else {
		err = r.Update(ctx, sa)
	}
	return err
}

func (r *ServiceReconciler) ensureRole(ctx context.Context, role *rbacv1.Role) error {

	name := types.NamespacedName{
		Name:      role.Name,
		Namespace: role.Namespace,
	}

	var tmp rbacv1.Role
	err := r.UncachedClient.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, role)
	} else {
		err = r.Update(ctx, role)
	}
	return err
}

func (r *ServiceReconciler) ensureRoleBinding(ctx context.Context, rb *rbacv1.RoleBinding) error {

	name := types.NamespacedName{
		Name:      rb.Name,
		Namespace: rb.Namespace,
	}

	var tmp rbacv1.RoleBinding
	err := r.UncachedClient.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, rb)
	} else {
		err = r.Update(ctx, rb)
	}
	return err
}

func (r *ServiceReconciler) deleteConfigMap(ctx context.Context, namespacedName types.NamespacedName) error {
	var tmp corev1.ConfigMap
	err := r.UncachedClient.Get(ctx, namespacedName, &tmp)
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
	err := r.UncachedClient.Get(ctx, namespacedName, &tmp)
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

func (r *ServiceReconciler) deleteServiceAccount(ctx context.Context, namespacedName types.NamespacedName) error {
	var tmp corev1.ServiceAccount
	err := r.UncachedClient.Get(ctx, namespacedName, &tmp)
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

func (r *ServiceReconciler) deleteRole(ctx context.Context, namespacedName types.NamespacedName) error {
	var tmp rbacv1.Role
	err := r.UncachedClient.Get(ctx, namespacedName, &tmp)
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

func (r *ServiceReconciler) deleteRoleBinding(ctx context.Context, namespacedName types.NamespacedName) error {
	var tmp rbacv1.RoleBinding
	err := r.UncachedClient.Get(ctx, namespacedName, &tmp)
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
