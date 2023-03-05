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

var (
	tsClient = tailscale.LocalClient{
		Socket: "/tmp/tailscaled.sock",
	}

	ctx = context.Background()

	sleepTime = 5 * time.Second

	k8sClient client.Client

	podName = os.Getenv("HOSTNAME")
)

func init() {
	var err error
	k8sClient, err = client.New(config.GetConfigOrDie(), client.Options{})
	errcheck.Check(err)
}

func main() {
	for {
		state, err := tsClient.Status(ctx)
		errcheck.Check(err)

		var tsip4 string

		for _, ip := range state.TailscaleIPs {
			if ip.Is4() {
				tsip4 = ip.String()
			}
		}

		pod := &corev1.Pod{}

		k8sClient.Get(ctx, types.NamespacedName{
			Name:      podName,
			Namespace: "tailscale",
		}, pod)

		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		if pod.Annotations["operator.pthomison.com/tailscale-ip"] != tsip4 {
			pod.Annotations["operator.pthomison.com/tailscale-ip"] = tsip4
			err := k8sClient.Update(ctx, pod)
			if err != nil {
				fmt.Println(err)
			}
		}

		time.Sleep(sleepTime)
	}
}
