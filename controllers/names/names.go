package names

import (
	"fmt"
	"os"

	"github.com/pthomison/errcheck"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

func TLBNamespace() string {
	return "tailscale"
}

func TLBNamespacedName(req *ctrl.Request) (string, string, types.NamespacedName) {
	name := fmt.Sprintf("tlb-%s-%s", req.Namespace, req.Name)
	namespace := TLBNamespace()
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	return name, namespace, namespacedName
}

func SelectorLabels(svcName string, svcNamespace string) (map[string]string, labels.Selector) {

	labelMap := make(map[string]string)

	labelMap[CommonLabel] = CommonLabelVal
	labelMap[ServiceNameLabel] = svcName
	labelMap[ServiceNamespaceLabel] = svcNamespace

	selector, err := labels.Parse(fmt.Sprintf("%s==%s,%s==%s,%s==%s", CommonLabel, CommonLabelVal, ServiceNameLabel, svcName, ServiceNamespaceLabel, svcNamespace))
	errcheck.Check(err)

	return labelMap, selector
}

func TLBImage() string {

	image := os.Getenv("TLB_IMAGE")
	if image == "" {
		image = "registry.localhost:15000/tailscale-lb:latest"
	}
	return image
}
