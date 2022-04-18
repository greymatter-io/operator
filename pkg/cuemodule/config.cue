// Contains user-configurable values. Editing this file will affect the way the operator is deployed and configured.
// Unlike inputs.cue, values will not be _injected_ into these fields from Go. If you set it here, that's it.

package only

config: {
  // Flags
  spire: bool | *true // enable Spire-based mTLS DEBUG - the default should be false
  auto_apply_mesh: bool | *true // apply the default mesh specified above after a delay // TODO, not actually used yet - implement
  generate_webhook_certs: bool | *true

  // Values
  cluster_ingress_name: "cluster" // For OpenShift deployments, this is used to look up the configured ingress domain
}
