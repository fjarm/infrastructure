package main

import (
	"github.com/fjarm/infrastructure/cert-manager/v1/pkg/release"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		err := release.DeployCertManager(ctx)
		return err
	})
}
