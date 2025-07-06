package dragonfly

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	operatorAppLabel     = "dragonfly-operator"
	operatorChartName    = "dragonfly-operator"
	operatorChartRepo    = "oci://ghcr.io/dragonflydb/dragonfly-operator/helm/dragonfly-operator"
	operatorChartVersion = "v1.1.11"
)

// DeployDragonflyOperatorHelmChart deploys the DragonflyDB operator using its Helm chart that can be found on GitHub
// packages.
//
// The return value is an array of the installed dependencies.
func DeployDragonflyOperatorHelmChart(
	ctx *pulumi.Context,
	provider *kubernetes.Provider,
	deps []pulumi.Resource,
) ([]pulumi.Resource, error) {
	dragonflyOperatorNamespace, err := deployDragonflyOperatorNamespace(ctx, provider)
	if err != nil {
		return nil, err
	}
	chartArgs := newDragonflyOperatorHelmChartArgs(dragonflyOperatorNamespace)
	chart, err := helmv4.NewChart(
		ctx,
		operatorChartName,
		chartArgs,
		pulumi.Provider(provider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}

	dragonflyClusterNamespace, err := deployDragonflyClusterNamespace(ctx, provider)
	if err != nil {
		return nil, err
	}
	dragonflyCluster, err := deployDragonflyCluster(
		ctx,
		dragonflyClusterNamespace,
		provider,
		[]pulumi.Resource{chart},
	)
	if err != nil {
		return nil, err
	}

	ctx.Export("dragonflyOperatorNamespace", dragonflyOperatorNamespace.Metadata.Name())

	return []pulumi.Resource{dragonflyOperatorNamespace, chart, dragonflyClusterNamespace, dragonflyCluster}, nil
}

// newDragonflyOperatorHelmChartArgs returns the Helm chart args used to deploy the Dragonfly Operator using the Helm
// chart.
func newDragonflyOperatorHelmChartArgs(
	namespace *corev1.Namespace,
) *helmv4.ChartArgs {
	chartArgs := &helmv4.ChartArgs{
		Chart:     pulumi.String(operatorChartRepo),
		Namespace: namespace.Metadata.Name(),
		Version:   pulumi.String(operatorChartVersion),
		Values: pulumi.Map{
			"replicaCount": pulumi.Int(3),
			"crds": pulumi.Map{
				"install": pulumi.Bool(true),
				"keep":    pulumi.Bool(false),
			},
			"manager": pulumi.Map{
				"resources": pulumi.Map{
					"limits": pulumi.Map{
						"cpu":    pulumi.String("500m"),
						"memory": pulumi.String("128Mi"),
					},
					"requests": pulumi.Map{
						"cpu":    pulumi.String("10m"),
						"memory": pulumi.String("64Mi"),
					},
				},
			},
		},
	}
	return chartArgs
}
