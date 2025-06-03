package main

import (
	"github.com/fjarm/infrastructure/pkg/v1/certmanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		err := certmanager.DeployCertManager(ctx)
		//err := VerifyCertManagerDeployed()
		//if err != nil {
		//	return err
		//}
		//err = dragonfly.DeployDragonfly(ctx)
		return err
	})
}
