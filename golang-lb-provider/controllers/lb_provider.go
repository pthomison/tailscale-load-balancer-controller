package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LoadBalancer struct {
	svc *corev1.Service
	LoadBalancerObjects
}

type LoadBalancerObjects struct {
	Deployment *appsv1.Deployment
	ConfigMap  *corev1.ConfigMap
}

func (lb *LoadBalancer) Render() {
	lb.renderDeployment()
	lb.renderConfigMap()
}

func (lb *LoadBalancer) Inject(r *ServiceReconciler, ctx context.Context) error {
	err := lb.ensureConfigMap(r, ctx)
	if err != nil {
		return err
	}

	err = lb.ensureDeployment(r, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (lb *LoadBalancer) ensureConfigMap(r *ServiceReconciler, ctx context.Context) error {

	name := types.NamespacedName{
		Name:      lb.LoadBalancerObjects.Deployment.ObjectMeta.Name,
		Namespace: lb.LoadBalancerObjects.Deployment.ObjectMeta.Namespace,
	}

	var tmp corev1.ConfigMap
	err := r.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, lb.LoadBalancerObjects.ConfigMap)
	} else {
		err = r.Update(ctx, lb.LoadBalancerObjects.ConfigMap)
	}
	return err
}

func (lb *LoadBalancer) ensureDeployment(r *ServiceReconciler, ctx context.Context) error {

	name := types.NamespacedName{
		Name:      lb.LoadBalancerObjects.Deployment.ObjectMeta.Name,
		Namespace: lb.LoadBalancerObjects.Deployment.ObjectMeta.Namespace,
	}

	var tmp appsv1.Deployment
	err := r.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, lb.LoadBalancerObjects.Deployment)
	} else {
		err = r.Update(ctx, lb.LoadBalancerObjects.Deployment)
	}
	return err
}
