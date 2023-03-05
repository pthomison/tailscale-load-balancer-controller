package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pthomison/errcheck"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"tailscale.com/client/tailscale"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	sleepTime        = 5 * time.Second
	tsSocketLocation = "/tmp/tailscaled.sock"
	ipAnnotation     = "operator.pthomison.com/tailscale-ip"
)

var (
	ctx          context.Context
	k8s          client.Client
	podName      string
	podNamespace string
	tailscaleIP  string
)

func init() {
	ctx = context.Background()

	var err error
	k8s, err = client.New(config.GetConfigOrDie(), client.Options{})
	errcheck.Check(err)

	podName = os.Getenv("HOSTNAME")
	podNamespace = "tailscale"
}

func main() {
	for {
		_, err := os.Stat(tsSocketLocation)
		if err == nil {
			break
		}
		fmt.Println("Waiting for Tailscale socket")
		time.Sleep(1 * time.Second)
	}

	tsClient := tailscale.LocalClient{
		Socket: tsSocketLocation,
	}

	for {
		state, err := tsClient.Status(ctx)
		errcheck.Check(err)

		for _, ip := range state.TailscaleIPs {
			if ip.Is4() {
				ips := ip.String()

				if tailscaleIP != ips {
					tailscaleIP = ips
					fmt.Printf("Tailscale IP Detected: %s\n", tailscaleIP)
				}
			}
		}

		pod := &corev1.Pod{}

		k8s.Get(ctx, types.NamespacedName{
			Name:      podName,
			Namespace: podNamespace,
		}, pod)

		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		if pod.Annotations[ipAnnotation] != tailscaleIP {
			fmt.Printf("Updating Pod Annotation: %s==%s\n", ipAnnotation, tailscaleIP)

			pod.Annotations[ipAnnotation] = tailscaleIP
			err := k8s.Update(ctx, pod)
			if err != nil {
				fmt.Println(err)
			}
		}

		time.Sleep(sleepTime)
	}
}
