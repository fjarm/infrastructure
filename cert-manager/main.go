package main

import (
	"github.com/fjarm/infrastructure/cert-manager/v1/pkg/release"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	configKind                   = "kind"
	exportCertManagerNamespace   = "certManagerNamespace"
	exportCertManagerStatus      = "certManagerStatus"
	helmReleaseName              = "cert-manager"
	k8sProviderLogicalNamePrefix = "kubernetes"
	valuesFilePath               = "v1/deploy/release/cert-manager-values.yaml"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		kind := config.GetBool(ctx, configKind)
		k8sProvider, err := kubernetes.NewProvider(ctx, k8sProviderLogicalNamePrefix, nil)
		if err != nil {
			return err
		}
		releaseArgs, err := release.NewCertManagerHelmReleaseArgs(kind, valuesFilePath)
		if err != nil {
			return err
		}
		certManager, err := helmv3.NewRelease(ctx, helmReleaseName, releaseArgs, pulumi.Provider(k8sProvider))
		if err != nil {
			return err
		}

		ctx.Export(exportCertManagerNamespace, certManager.Namespace)
		ctx.Export(exportCertManagerStatus, certManager.Status)
		return nil
	})
}
