package controllers

import (
	"context"
	"fmt"

	"github.com/pthomison/tailscale-load-balancer-controller/controllers/lb"
	"github.com/pthomison/tailscale-load-balancer-controller/controllers/names"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ServiceReconciler) Inject(ctx context.Context, LB *lb.LoadBalancer) error {
	err := r.ensureConfigMap(ctx, LB.ConfigMap)
	if err != nil {
		return err
	}

	err = r.ensureDeployment(ctx, LB.Deployment)
	if err != nil {
		return err
	}

	return nil
}

func Delete(r *ServiceReconciler, ctx context.Context, req *ctrl.Request) error {
	_, _, dpNamespacedName := names.TLBDeploymentName(req)
	err := r.deleteDeployment(ctx, dpNamespacedName)
	if err != nil {
		return err
	}

	_, _, cmNamespacedName := names.TLBConfigMapName(req)
	err = r.deleteConfigMap(ctx, cmNamespacedName)
	if err != nil {
		return err
	}

	return nil
}

func CheckForOrphanedDeplyments(r *ServiceReconciler, ctx context.Context) error {
	ns := names.TLBNamespace()

	fmt.Println("Check For Ophans")

	var deploymentList appsv1.DeploymentList
	err := r.List(ctx, &deploymentList, client.InNamespace(ns), client.MatchingLabels{
		names.CommonLabel: names.CommonLabelVal,
	})
	if client.IgnoreNotFound(err) != nil {
		return err
	}

	for _, deployment := range deploymentList.Items {
		svcName := deployment.Labels[names.ServiceNameLabel]
		svcNamespace := deployment.Labels[names.ServiceNamespaceLabel]

		fmt.Printf("Existing LB Detected: %s/%s\n", svcName, svcNamespace)

		exists, _, err := r.getService(ctx, svcName, svcNamespace)
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
