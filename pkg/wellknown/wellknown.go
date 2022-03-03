package wellknown

const (
	ANNOTATION_INJECT_SIDECAR_TO_PORT = "greymatter.io/inject-sidecar-to" // whether to inject sidecar, and upstream port
	ANNOTATION_CONFIGURE_SIDECAR      = "greymatter.io/configure-sidecar" // whether to apply automatic configuration to sidecar
	ANNOTATION_LAST_APPLIED           = "greymatter.io/last-applied"
	LABEL_CLUSTER                     = "greymatter.io/cluster"
	LABEL_WORKLOAD                    = "greymatter.io/workload"
)
