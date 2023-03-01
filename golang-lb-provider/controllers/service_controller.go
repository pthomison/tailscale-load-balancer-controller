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

	"github.com/pthomison/errcheck"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
)

var (
	defaultNamespace = "tailscale"
	defaultSecret    = "tailscale-token"
	defaultSecretKey = "token"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;create;update;delete;watch;patch
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;create;update;delete;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;create;update;delete;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Request the service
	var svc corev1.Service
	err := r.Get(ctx, req.NamespacedName, &svc)
	if client.IgnoreNotFound(err) != nil {
		// requeue in hopes that the error is transient
		return ctrl.Result{}, err
	} else if err != nil {
		// if object has been deleted, ignore
		return ctrl.Result{}, nil
	}

	if svc.Spec.Type != "LoadBalancer" {
		// non LB service, ignore
		return ctrl.Result{}, nil
	}

	fmt.Printf("Service: %v/%v/%v\n", svc.Spec.Type, svc.Name, svc.Namespace)

	lb := LoadBalancer{
		svc: &svc,
	}

	lb.Render()
	err = lb.Inject(r, ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	var lbPodList corev1.PodList
	for {
		_, selector := SelectorLabels(lb.svc)

		for len(lbPodList.Items) == 0 {
			err = r.List(ctx, &lbPodList, &client.ListOptions{
				LabelSelector: client.MatchingLabelsSelector{
					Selector: selector,
				},
			})
			errcheck.Check(err)

			time.Sleep(1 * time.Second)
		}

		pod := lbPodList.Items[0]

		fmt.Println(pod.Annotations)

		break

	}

	fmt.Printf("%v\n", lbPodList.Items)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}
