package certmanager

import (
	"fmt"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
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
	helmChartName              = "cert-manager"
)

// DeployCertManager deploys the cert-manager Helm chart.
func DeployCertManager(ctx *pulumi.Context, provider *kubernetes.Provider) ([]pulumi.Resource, error) {
	kind := config.GetBool(ctx, configKind)

	ns, err := newCertManagerNamespace(ctx)
	if err != nil {
		return nil, err
	}

	chartArgs := newCertManagerHelmChartArgs(ns, kind)
	certManager, err := helmv4.NewChart(ctx, helmChartName, chartArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	//clusterIssuer, err := DeployCertManagerInternalClusterIssuer(ctx, provider, []pulumi.Resource{certManager}, kind)
	//if err != nil {
	//	return nil, err
	//}

	ctx.Export(exportCertManagerNamespace, ns)
	ctx.Export(exportCertManagerStatus, certManager)
	return []pulumi.Resource{
		certManager,
		//clusterIssuer,
	}, nil
}

// newCertManagerHelmChartArgs creates a Helm chart arguments. The Helm chart args can then be used by a Pulumi program
// to deploy cert-manager.
//
// [kind] controls the [enableGatewayAPI] value by disable the GatewayAPI if the chart is deployed locally.
func newCertManagerHelmChartArgs(ns *corev1.Namespace, kind bool) *helmv4.ChartArgs {
	chartArgs := &helmv4.ChartArgs{
		Chart: pulumi.String(chartName),
		RepositoryOpts: &helmv4.RepositoryOptsArgs{
			Repo: pulumi.String(chartRepo),
		},
		Namespace: ns.Metadata.Name(),
		Version:   pulumi.String(chartVersion),
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
				"config": pulumi.Map{
					"apiVersion": pulumi.String("webhook.config.cert-manager.io/v1alpha1"),
					"kind":       pulumi.String("WebhookConfiguration"),
					"logging": pulumi.Map{
						"format":    pulumi.String("json"),
						"verbosity": pulumi.Int(5), // Debug
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
		},
	}
	return chartArgs
}

// newCertManagerNamespace creates a Namespace for cert-manager to be installed in.
func newCertManagerNamespace(ctx *pulumi.Context) (*corev1.Namespace, error) {
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
