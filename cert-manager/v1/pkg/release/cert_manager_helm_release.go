package release

import (
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	chartName      = "cert-manager"
	chartNamespace = "cert-manager"
	chartRepo      = "https://charts.jetstack.io"
	chartVersion   = "1.17.2"
)

func NewCertManagerHelmReleaseArgs() (*helmv3.ReleaseArgs, error) {
	releaseArgs := &helmv3.ReleaseArgs{
		Chart: pulumi.String(chartName),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(chartRepo),
		},
		Namespace:       pulumi.String(chartNamespace),
		Version:         pulumi.String(chartVersion),
		Atomic:          pulumi.Bool(true),
		CreateNamespace: pulumi.Bool(true),
		DisableCRDHooks: pulumi.Bool(false),
		Timeout:         pulumi.Int(120),
		ValueYamlFiles: pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset("cert-manager-values.yaml"),
		},
	}
	return releaseArgs, nil
}
