package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pthomison/errcheck"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn/ipnstate"

	corev1 "k8s.io/api/core/v1"
)

const (
	sleepTime        = 5 * time.Second
	tsSocketLocation = "/tmp/tailscaled.sock"
	ipAnnotation     = "operator.pthomison.com/tailscale-ip"
	dnsAnnotation    = "operator.pthomison.com/tailscale-dns"
)

var (
	ctx          context.Context
	k8s          client.Client
	podName      string
	podNamespace string
	tailscaleIP  string
	tailscaleDNS string
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
		// Capture Pod
		pod := &corev1.Pod{}
		k8s.Get(ctx, types.NamespacedName{
			Name:      podName,
			Namespace: podNamespace,
		}, pod)

		// Capture Tailscale Status
		state, err := tsClient.Status(ctx)
		errcheck.Check(err)

		// Check IPV4 Address & Update If Needed
		currentIP := captureIP4(state)
		if tailscaleIP != currentIP {
			tailscaleIP = currentIP
			fmt.Printf("Tailscale IP Detected: %s\n", tailscaleIP)

			pod = updateAnnotation(pod, ipAnnotation, tailscaleIP)
		}

		// Check Tailscale MagicDNS & Update If Needed
		currentDNS := state.Self.DNSName
		if tailscaleDNS != currentDNS {
			tailscaleDNS = currentDNS
			fmt.Printf("Tailscale DNS Detected: %s\n", tailscaleDNS)

			tailscaleDNS = strings.Trim(tailscaleDNS, ".")

			pod = updateAnnotation(pod, dnsAnnotation, tailscaleDNS)
		}

		_ = pod

		time.Sleep(sleepTime)
	}
}

func updateAnnotation(pod *corev1.Pod, annotationKey string, annotationValue string) *corev1.Pod {
	fmt.Printf("Updating Pod (%s) Annotation: %s==%s\n", pod.Name, annotationKey, annotationValue)

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}

	pod.Annotations[annotationKey] = annotationValue
	err := k8s.Update(ctx, pod)
	if err != nil {
		fmt.Println(err)
	}
	return pod
}

func captureIP4(state *ipnstate.Status) string {
	for _, ip := range state.TailscaleIPs {
		if ip.Is4() {
			return ip.String()
		}
	}
	return ""
}
