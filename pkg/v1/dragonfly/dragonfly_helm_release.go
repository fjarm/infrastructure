package dragonfly

import (
	"fmt"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"time"
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
			"extraArgs": pulumi.StringArray{
				pulumi.String("--cluster_mode=emulated"),
				pulumi.String("--admin_port=8000"),
				pulumi.String(fmt.Sprintf("--dbfilename=my-dump-%s}", time.Now().Format(time.RFC3339))),
				pulumi.String("--snapshot_cron=* * * * *"), // Snapshot every minute
			},
			"podSecurityContext": pulumi.Map{
				"fsGroup": pulumi.Int(2000),
			},
			"securityContext": pulumi.Map{
				"capabilities": pulumi.Map{
					"drop": pulumi.StringArray{
						pulumi.String("ALL"),
					},
				},
				"readOnlyRootFileSystem": pulumi.Bool(true),
				"runAsNonRoot":           pulumi.Bool(true),
				"runAsUser":              pulumi.Int(1000),
			},
			"tls": pulumi.Map{
				"enabled":     pulumi.Bool(true),
				"createCerts": pulumi.Bool(true),
				"issuer": pulumi.Map{
					"kind": pulumi.String("ClusterIssuer"),
					"name": pulumi.String("default"),
				},
			},
			"replicaCount": pulumi.Int(1),
			"resources": pulumi.Map{
				"limits": pulumi.Map{
					"memory": pulumi.String("2Gi"),
				},
			},
		},
	}
	return releaseArgs
}
