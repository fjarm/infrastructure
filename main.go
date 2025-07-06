package main

import (
	"github.com/fjarm/infrastructure/pkg/v1/certmanager"
	"github.com/fjarm/infrastructure/pkg/v1/valkey"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	k8sProviderLogicalNamePrefix = "kubernetes"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		k8sProvider, err := kubernetes.NewProvider(ctx, k8sProviderLogicalNamePrefix, &kubernetes.ProviderArgs{
			//RenderYamlToDirectory: pulumi.String("yaml"),
		})
		if err != nil {
			return err
		}
		certManagerDeps, err := certmanager.DeployCertManager(
			ctx,
			k8sProvider,
		)
		if err != nil {
			return err
		}

		_, err = valkey.DeployValkeyCluster(
			ctx,
			k8sProvider,
			certManagerDeps,
		)
		return err
	})
}
