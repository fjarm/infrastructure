package dragonfly

import (
	"fmt"
	"github.com/fjarm/infrastructure/pkg/v1/certmanager"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"time"
)

const (
	chartRepo                = "oci://ghcr.io/dragonflydb/dragonfly/helm/dragonfly"
	chartNamespace           = "dragonfly"
	chartVersion             = "v1.30.3"
	exportDragonflyNamespace = "dragonflyNamespace"
	exportDragonflyStatus    = "dragonflyStatus"
	helmChartName            = "dragonfly"
)

// DeployDragonfly attempts to deploy DragonflyDB using the official Helm chart.
func DeployDragonfly(ctx *pulumi.Context, provider *kubernetes.Provider, deps []pulumi.Resource) error {
	ns, err := newDragonflyNamespace(ctx)
	if err != nil {
		return err
	}

	chartArgs := newDragonflyHelmChartArgs(ns)
	dragonflyChart, err := helmv4.NewChart(
		ctx,
		helmChartName,
		chartArgs,
		pulumi.Provider(provider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return err
	}

	ctx.Export(exportDragonflyNamespace, ns)
	ctx.Export(exportDragonflyStatus, dragonflyChart)
	return nil
}

// newDragonflyHelmChartArgs constructs the Helm chart values needed to deploy DragonflyDB to k8s.
func newDragonflyHelmChartArgs(ns *corev1.Namespace) *helmv4.ChartArgs {
	chartArgs := &helmv4.ChartArgs{
		Chart:     pulumi.String(chartRepo),
		Namespace: ns.Metadata.Name(),
		Version:   pulumi.String(chartVersion),
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
				// TODO(2025-06-06): Update this to be more secure
				pulumi.String("--requirepass=password"),
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
					"name": pulumi.String(certmanager.InternalClusterIssuerName),
				},
				"cert": pulumi.String("tls.crt"),
				"key":  pulumi.String("tls.key"),
			},
			"replicaCount": pulumi.Int(1),
			"resources": pulumi.Map{
				"limits": pulumi.Map{
					"memory": pulumi.String("2Gi"),
				},
			},
		},
	}
	return chartArgs
}

// newDragonflyNamespace creates a Namespace for DragonflyDB to be installed in.
func newDragonflyNamespace(ctx *pulumi.Context) (*corev1.Namespace, error) {
	ns, err := corev1.NewNamespace(ctx, chartNamespace, &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(chartNamespace),
			Labels: pulumi.StringMap{
				"app": pulumi.String(helmChartName),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return ns, nil
}
