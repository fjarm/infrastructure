package certmanager

import (
	"fmt"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	chartName                  = "cert-manager"
	chartNamespace             = "cert-manager"
	chartRepo                  = "https://charts.jetstack.io"
	chartVersion               = "1.17.2"
	configKind                 = "certmanager:kind"
	exportCertManagerNamespace = "certManagerNamespace"
	exportCertManagerStatus    = "certManagerStatus"
	helmReleaseName            = "cert-manager"
)

// DeployCertManager deploys the cert-manager Helm chart.
func DeployCertManager(ctx *pulumi.Context, provider *kubernetes.Provider) ([]pulumi.Resource, error) {
	kind := config.GetBool(ctx, configKind)

	releaseArgs := NewCertManagerHelmReleaseArgs(kind)
	certManager, err := helmv3.NewRelease(ctx, helmReleaseName, releaseArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	clusterIssuer, err := DeployCertManagerInternalClusterIssuer(ctx, provider, []pulumi.Resource{certManager}, kind)
	if err != nil {
		return nil, err
	}

	ctx.Export(exportCertManagerNamespace, certManager.Namespace)
	ctx.Export(exportCertManagerStatus, certManager.Status)
	return []pulumi.Resource{certManager, clusterIssuer}, nil
}

// NewCertManagerHelmReleaseArgs creates a Helm release with values that match 1 for 1 the cert-manager-values.yaml
// file. The Helm release can then be used by a Pulumi program to deploy cert-manager.
//
// [kind] controls the [enableGatewayAPI] value by disable the GatewayAPI if the chart is deployed locally.
func NewCertManagerHelmReleaseArgs(kind bool) *helmv3.ReleaseArgs {
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
		Values: pulumi.Map{
			"global": pulumi.Map{
				"leaderElection": pulumi.Map{
					"namespace": pulumi.String(chartNamespace),
				},
			},
			"image": pulumi.Map{
				"tag": pulumi.String(fmt.Sprintf("v%s", chartVersion)),
			},
			"replicaCount": pulumi.Int(3),
			"podDisruptionBudget": pulumi.Map{
				"enabled": pulumi.Bool(true),
			},
			"strategy": pulumi.Map{
				"type": pulumi.String("RollingUpdate"),
				"rollingUpdate": pulumi.Map{
					"maxSurge":       pulumi.Int(1),
					"maxUnavailable": pulumi.Int(1),
				},
			},
			"config": pulumi.Map{
				"apiVersion": pulumi.String("controller.config.cert-manager.io/v1alpha1"),
				"kind":       pulumi.String("ControllerConfiguration"),
				"logging": pulumi.Map{
					"format":    pulumi.String("json"),
					"verbosity": pulumi.Int(5), // Debug
				},
				"leaderElectionConfig": pulumi.Map{
					"namespace": pulumi.String(chartNamespace),
				},
				"enableGatewayAPI": pulumi.Bool(!kind),
			},
			"crds": pulumi.Map{
				"enabled": pulumi.Bool(true),
				"keep":    pulumi.Bool(false),
			},
			"prometheus": pulumi.Map{
				"enabled": pulumi.Bool(true),
			},
			"cainjector": pulumi.Map{
				"replicaCount": pulumi.Int(3),
				"config": pulumi.Map{
					"apiVersion": pulumi.String("cainjector.config.cert-manager.io/v1alpha1"),
					"kind":       pulumi.String("CAInjectorConfiguration"),
					"logging": pulumi.Map{
						"format":    pulumi.String("json"),
						"verbosity": pulumi.Int(5), // Debug
					},
					"leaderElectionConfig": pulumi.Map{
						"namespace": pulumi.String(chartNamespace),
					},
				},
				"strategy": pulumi.Map{
					"type": pulumi.String("RollingUpdate"),
					"rollingUpdate": pulumi.Map{
						"maxSurge":       pulumi.Int(0),
						"maxUnavailable": pulumi.Int(1),
					},
				},
				"podDisruptionBudget": pulumi.Map{
					"enabled": pulumi.Bool(true),
				},
			},
			"webhook": pulumi.Map{
				"replicaCount": pulumi.Int(3),
				"strategy": pulumi.Map{
					"type": pulumi.String("RollingUpdate"),
					"rollingUpdate": pulumi.Map{
						"maxSurge":       pulumi.Int(0),
						"maxUnavailable": pulumi.Int(1),
					},
				},
				"podDisruptionBudget": pulumi.Map{
					"enabled": pulumi.Bool(true),
				},
			},
		},
	}
	return releaseArgs
}
