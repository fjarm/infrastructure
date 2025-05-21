package release

import (
	"context"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
)

const (
	chartName      = "cert-manager"
	chartNamespace = "cert-manager"
	chartRepo      = "https://charts.jetstack.io"
	chartVersion   = "1.17.2"
)

func NewCertManagerHelmRelease(ctx context.Context, k8sProvider *kubernetes.Provider) (*helmv3.Release, error) {
	return nil, ErrUnimplemented
}
