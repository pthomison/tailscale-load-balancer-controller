package names

const (
	annotationBase = "operator.pthomison.com"

	serviceNameLabel      = "operator.pthomison.com/service-name-ref"
	serviceNamespaceLabel = "operator.pthomison.com/service-namespace-ref"
	commonLabel           = "app.kubernetes.io/name"
	commonLabelVal        = "tailscale-lb-provider"

	DefaultSecret    = "tailscale-token"
	DefaultSecretKey = "token"
)
