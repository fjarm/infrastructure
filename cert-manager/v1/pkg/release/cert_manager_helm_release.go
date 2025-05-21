package release

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	chartName      = "cert-manager"
	chartNamespace = "cert-manager"
	chartRepo      = "https://charts.jetstack.io"
	chartVersion   = "1.17.2"
)

func NewCertManagerHelmRelease(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) (*helmv3.Release, error) {
	return nil, ErrUnimplemented
}
