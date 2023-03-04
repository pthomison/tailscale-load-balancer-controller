package names

const (
	AnnotationBase = "operator.pthomison.com"

	ServiceNameLabel      = "operator.pthomison.com/service-name-ref"
	ServiceNamespaceLabel = "operator.pthomison.com/service-namespace-ref"
	CommonLabel           = "app.kubernetes.io/name"
	CommonLabelVal        = "tailscale-lb-provider"

	DefaultSecret    = "tailscale-token"
	DefaultSecretKey = "token"
)
