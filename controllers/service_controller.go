/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
)

var (
	startup = true
)

const (
	annotationBase = "operator.pthomison.com"

	serviceNameLabel      = "operator.pthomison.com/service-name-ref"
	serviceNamespaceLabel = "operator.pthomison.com/service-namespace-ref"
	commonLabel           = "app.kubernetes.io/name"
	commonLabelVal        = "tailscale-lb-provider"

	defaultSecret    = "tailscale-token"
	defaultSecretKey = "token"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;create;update;delete;watch;patch
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;create;update;delete;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;create;update;delete;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	var err error

	if startup {
		err = CheckForOrphanedDeplyments(r, ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
		startup = false
	}

	// Request the service
	exists, svc, err := r.GetService(ctx, req.Name, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	if exists {

		if svc.Spec.Type != "LoadBalancer" {
			// non LB service, ignore
			return ctrl.Result{}, nil
		}

		err = r.EnsureLoadBalancer(ctx, svc, req)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		err = r.DestroyLoadBalancer(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}

func (r *ServiceReconciler) EnsureLoadBalancer(ctx context.Context, svc *corev1.Service, req ctrl.Request) error {
	fmt.Printf("Service: %v/%v/%v\n", svc.Spec.Type, svc.Name, svc.Namespace)

	lb := LoadBalancer{
		req: &req,
		svc: svc,
	}

	lb.Render()
	err := lb.Inject(r, ctx)
	if err != nil {
		return err
	}

	var lbPodList corev1.PodList
	var loadbalancerIP string
	for {
		_, selector := SelectorLabels(lb.req.Name, lb.req.Namespace)

		err = r.List(ctx, &lbPodList, &client.ListOptions{
			LabelSelector: client.MatchingLabelsSelector{
				Selector: selector,
			},
		})
		if err != nil {
			return err
		}

		if len(lbPodList.Items) != 0 {
			pod := lbPodList.Items[0]

			annotation := fmt.Sprintf("%s/tailscale-ip", annotationBase)
			if pod.Annotations[annotation] != "" {
				loadbalancerIP = pod.Annotations[annotation]
				break
			}
		}
		fmt.Println("Waiting for tailscale IP")
		time.Sleep(5 * time.Second)
	}

	lb.svc.Spec.ExternalIPs = []string{loadbalancerIP}

	err = r.Update(ctx, lb.svc)
	if err != nil {
		return err
	}

	return nil
}

func (r *ServiceReconciler) DestroyLoadBalancer(ctx context.Context, req ctrl.Request) error {
	return Delete(r, ctx, &req)
}