package dragonfly

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	chartRepo                    = "oci://ghcr.io/dragonflydb/dragonfly/helm/dragonfly"
	chartNamespace               = "dragonfly"
	chartVersion                 = "v1.30.3"
	exportDragonflyNamespace     = "dragonflyNamespace"
	exportDragonflyStatus        = "dragonflyStatus"
	helmReleaseName              = "dragonfly"
	k8sProviderLogicalNamePrefix = "kubernetes"
)

// DeployDragonfly attempts to deploy DragonflyDB using the official Helm chart.
func DeployDragonfly(ctx *pulumi.Context) error {
	k8sProvider, err := kubernetes.NewProvider(ctx, k8sProviderLogicalNamePrefix, nil)
	if err != nil {
		return err
	}

	releaseArgs := NewDragonflyHelmReleaseArgs()
	dragonflyRelease, err := helmv3.NewRelease(ctx, helmReleaseName, releaseArgs, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	ctx.Export(exportDragonflyNamespace, dragonflyRelease.Namespace)
	ctx.Export(exportDragonflyStatus, dragonflyRelease.Status)
	return nil
}

// NewDragonflyHelmReleaseArgs constructs the Helm chart values needed to deploy DragonflyDB to k8s.
func NewDragonflyHelmReleaseArgs() *helmv3.ReleaseArgs {
	releaseArgs := &helmv3.ReleaseArgs{
		Chart:           pulumi.String(chartRepo),
		Namespace:       pulumi.String(chartNamespace),
		Version:         pulumi.String(chartVersion),
		Atomic:          pulumi.Bool(true),
		CreateNamespace: pulumi.Bool(true),
		DisableCRDHooks: pulumi.Bool(false),
		Timeout:         pulumi.Int(120),
		Values: pulumi.Map{
			"storage": pulumi.Map{
				"enabled":  pulumi.Bool(true),
				"requests": pulumi.String("2Gi"),
			},
			"replicaCount": pulumi.Int(1),
			"resources": pulumi.Map{
				"limits": pulumi.Map{
					"memory": pulumi.String("2Gi"),
				},
			},
			"extraArgs": pulumi.StringArray{
				pulumi.String("--cluster_mode=emulated"),
				pulumi.String("--admin_port=8000"),
			},
		},
	}
	return releaseArgs
}
