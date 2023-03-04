package controllers

import (
	"context"
	"fmt"

	"github.com/pthomison/tailscale-load-balancer-controller/controllers/lb"
	"github.com/pthomison/tailscale-load-balancer-controller/controllers/names"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// type LoadBalancer struct {
// 	req *ctrl.Request
// 	svc *corev1.Service
// 	LoadBalancerObjects
// }

// type LoadBalancerObjects struct {
// 	Deployment *appsv1.Deployment
// 	ConfigMap  *corev1.ConfigMap
// }

func (r *ServiceReconciler) Inject(ctx context.Context, LB *lb.LoadBalancer) error {
	err := r.ensureConfigMap(ctx, LB)
	if err != nil {
		return err
	}

	err = r.ensureDeployment(ctx, LB)
	if err != nil {
		return err
	}

	return nil
}

func Delete(r *ServiceReconciler, ctx context.Context, req *ctrl.Request) error {
	err := deleteDeployment(r, ctx, req)
	if err != nil {
		return err
	}

	err = deleteConfigMap(r, ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func deleteConfigMap(r *ServiceReconciler, ctx context.Context, req *ctrl.Request) error {
	_, _, namespacedName := names.TLBConfigMapName(req)

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

func deleteDeployment(r *ServiceReconciler, ctx context.Context, req *ctrl.Request) error {
	_, _, namespacedName := names.TLBDeploymentName(req)

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

func (r *ServiceReconciler) ensureConfigMap(ctx context.Context, LB *lb.LoadBalancer) error {

	name := types.NamespacedName{
		Name:      LB.Deployment.ObjectMeta.Name,
		Namespace: LB.Deployment.ObjectMeta.Namespace,
	}

	var tmp corev1.ConfigMap
	err := r.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, LB.ConfigMap)
	} else {
		err = r.Update(ctx, LB.ConfigMap)
	}
	return err
}

func (r *ServiceReconciler) ensureDeployment(ctx context.Context, LB *lb.LoadBalancer) error {

	name := types.NamespacedName{
		Name:      LB.Deployment.ObjectMeta.Name,
		Namespace: LB.Deployment.ObjectMeta.Namespace,
	}

	var tmp appsv1.Deployment
	err := r.Get(ctx, name, &tmp)
	if client.IgnoreNotFound(err) != nil {
		return err
	} else if err != nil {
		err = r.Create(ctx, LB.Deployment)
	} else {
		err = r.Update(ctx, LB.Deployment)
	}
	return err
}

func CheckForOrphanedDeplyments(r *ServiceReconciler, ctx context.Context) error {
	ns := names.TLBNamespace()

	fmt.Println("Check For Ophans")

	var deploymentList appsv1.DeploymentList
	err := r.List(ctx, &deploymentList, client.InNamespace(ns), client.MatchingLabels{
		commonLabel: commonLabelVal,
	})
	if client.IgnoreNotFound(err) != nil {
		return err
	}

	for _, deployment := range deploymentList.Items {
		svcName := deployment.Labels[serviceNameLabel]
		svcNamespace := deployment.Labels[serviceNamespaceLabel]

		fmt.Printf("Existing LB Detected: %s/%s\n", svcName, svcNamespace)

		exists, _, err := r.GetService(ctx, svcName, svcNamespace)
		if err != nil {
			return err
		}

		if !exists {
			fmt.Println("Service Does Not Exist; Terminating LB")
			err = r.DestroyLoadBalancer(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      svcName,
					Namespace: svcNamespace,
				},
			})
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Service Exists; Keeping LB in place")
		}

	}

	return nil
}
