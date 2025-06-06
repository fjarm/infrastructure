package main

import (
	"github.com/fjarm/infrastructure/pkg/v1/certmanager"
	"github.com/fjarm/infrastructure/pkg/v1/dragonfly"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	k8sProviderLogicalNamePrefix = "kubernetes"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		k8sProvider, err := kubernetes.NewProvider(ctx, k8sProviderLogicalNamePrefix, nil)
		if err != nil {
			return err
		}
		deps, err := certmanager.DeployCertManager(ctx, k8sProvider)
		if err != nil {
			return err
		}
		err = dragonfly.DeployDragonfly(ctx, k8sProvider, deps)
		return err
	})
}
