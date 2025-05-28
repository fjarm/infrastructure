package certmanager

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	chartName                    = "cert-manager"
	chartNamespace               = "cert-manager"
	chartRepo                    = "https://charts.jetstack.io"
	chartVersion                 = "1.17.2"
	configKind                   = "cert-manager:kind"
	exportCertManagerNamespace   = "certManagerNamespace"
	exportCertManagerStatus      = "certManagerStatus"
	helmReleaseName              = "cert-manager"
	k8sProviderLogicalNamePrefix = "kubernetes"
	valuesConfig                 = "config"
	valuesEnableGatewayAPI       = "enableGatewayAPI"
	valuesFilePath               = "v1/deploy/release/cert-manager-values.yaml"
)

func DeployCertManager(ctx *pulumi.Context) error {
	kind := config.GetBool(ctx, configKind)
	k8sProvider, err := kubernetes.NewProvider(ctx, k8sProviderLogicalNamePrefix, nil)
	if err != nil {
		return err
	}
	releaseArgs := NewCertManagerHelmReleaseArgs(kind, valuesFilePath)
	certManager, err := helmv3.NewRelease(ctx, helmReleaseName, releaseArgs, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	ctx.Export(exportCertManagerNamespace, certManager.Namespace)
	ctx.Export(exportCertManagerStatus, certManager.Status)
	return nil
}

func NewCertManagerHelmReleaseArgs(kind bool, valuesFilePath string) *helmv3.ReleaseArgs {
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
			pulumi.NewFileAsset(valuesFilePath),
		},
		Values: pulumi.Map{
			valuesConfig: pulumi.Map{
				valuesEnableGatewayAPI: pulumi.Bool(!kind),
			},
		},
	}
	return releaseArgs
}
