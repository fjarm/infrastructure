package release

import (
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
)

// VerifyCertManagerDeployed references the cert-manager Pulumi Stack to verify that the cert-manager Helm chart has
// been deployed. Without this chart, we cannot request TLS certificates.
func VerifyCertManagerDeployed() error {
	return ErrUnimplemented
}

// NewRedisHelmReleaseArgs constructs the Helm chart values needed to deploy Redis to k8s.
func NewRedisHelmReleaseArgs() *helmv3.ReleaseArgs {
	return nil
}
