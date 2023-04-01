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

	"github.com/pthomison/tailscale-load-balancer-controller/controllers/lb"
	"github.com/pthomison/tailscale-load-balancer-controller/controllers/names"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	StartUp        bool
	UncachedClient client.Client
}

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services/status,verbs=update

//+kubebuilder:rbac:namespace="system",groups="",resources=pods,verbs=get;list;update
//+kubebuilder:rbac:namespace="system",groups="",resources=secrets,verbs=get;create;update
//+kubebuilder:rbac:namespace="system",groups="",resources=configmaps,verbs=get;create;update;delete
//+kubebuilder:rbac:namespace="system",groups="",resources=serviceaccounts,verbs=get;create;update;delete
//+kubebuilder:rbac:namespace="system",groups="apps",resources=deployments,verbs=get;list;create;update;delete
// +kubebuilder:rbac:namespace="system",groups="rbac.authorization.k8s.io",resources=roles,verbs=get;create;update;delete
// +kubebuilder:rbac:namespace="system",groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	var err error

	if r.StartUp {
		err = CheckForOrphanedDeplyments(r, ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
		r.StartUp = false
	}

	// Request the service
	exists, svc, err := r.getService(ctx, req.Name, req.Namespace)
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

	LB := &lb.LoadBalancer{
		ServiceRequest: &req,
		Service:        svc,
	}

	LB.Render()
	err := r.Inject(ctx, LB)
	if err != nil {
		return err
	}

	_, selector := names.SelectorLabels(LB.ServiceRequest.Name, LB.ServiceRequest.Namespace)
	lbIP, lbDNS, err := WaitForTailscaleAnnotations(ctx, r.UncachedClient, selector, LB.Deployment.Namespace)
	if err != nil {
		return err
	}

	fmt.Printf("Discovered tailscaled IP: %v\n", lbIP)
	fmt.Printf("Discovered tailscaled DNS: %v\n", lbDNS)

	// LB.Service.Spec.ExternalIPs = []string{loadbalancerIP}

	// err = r.Update(ctx, LB.Service)
	// if err != nil {
	// 	return err
	// }

	lbIngress := corev1.LoadBalancerIngress{
		IP: lbIP,
	}
	if lbDNS != "" {
		lbIngress.Hostname = lbDNS
	}

	LB.Service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{lbIngress}

	err = r.Status().Update(ctx, LB.Service)
	if err != nil {
		return err
	}

	return nil
}

func (r *ServiceReconciler) DestroyLoadBalancer(ctx context.Context, req ctrl.Request) error {
	return Delete(r, ctx, &req)
}

func WaitForTailscaleAnnotations(ctx context.Context, k8sClient client.Client, podSelector labels.Selector, namespace string) (string, string, error) {
	var lbPodList corev1.PodList

	tailscaleIP := ""
	tailscaleDNS := ""
	var err error

	for {
		err := k8sClient.List(ctx, &lbPodList, &client.ListOptions{
			LabelSelector: client.MatchingLabelsSelector{
				Selector: podSelector,
			},
			Namespace: namespace,
		})
		if err != nil {
			return tailscaleIP, tailscaleDNS, err
		}

		if len(lbPodList.Items) != 0 {
			pod := lbPodList.Items[0]

			ipAnnotation := fmt.Sprintf("%s/tailscale-ip", names.AnnotationBase)
			dnsAnnotation := fmt.Sprintf("%s/tailscale-dns", names.AnnotationBase)
			if pod.Annotations[ipAnnotation] != "" {
				tailscaleIP = pod.Annotations[ipAnnotation]
			}
			if pod.Annotations[dnsAnnotation] != "" {
				tailscaleDNS = pod.Annotations[dnsAnnotation]
			}

			if tailscaleIP != "" {
				break
			}

		}
		fmt.Println("Waiting for tailscale initialization")
		time.Sleep(5 * time.Second)
	}

	return tailscaleIP, tailscaleDNS, err
}
