package valkey

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	clusterAppLabel   = "valkey"
	clusterAppVersion = "8.1.2"
	chartName         = "valkey"
	chartRepo         = "oci://registry-1.docker.io/bitnamicharts/valkey"
	chartVersion      = "3.0.16"
)

// DeployValkeyCluster sets up the required resources needed to run Valkey including a namespace, TLS certificate, and
// Helm chart.
func DeployValkeyCluster(
	ctx *pulumi.Context,
	provider *kubernetes.Provider,
	deps []pulumi.Resource,
) ([]pulumi.Resource, error) {
	namespace, err := deployValkeyClusterNamespace(
		ctx,
		provider,
		[]pulumi.Resource{},
	)
	if err != nil {
		return nil, err
	}

	cert, err := deployValkeyClusterCertificate(
		ctx,
		namespace,
		provider,
		append(deps, namespace),
	)
	if err != nil {
		return nil, err
	}

	aclContent, err := newValkeyUserACL(
		// TODO(2025-06-30): Add support to the template for creating users with no password
		&valkeyUser{
			Username:        "test",
			Password:        "test",
			EnabledCommands: []string{"+AUTH", "+PING", "+GET", "+SET", "~*"},
		},
		&valkeyUser{
			Username:        "default",
			Password:        "password",
			EnabledCommands: []string{"+@ALL", "~*"},
		},
	)
	if err != nil {
		return nil, err
	}

	chart, err := deployValkeyClusterHelmChart(
		ctx,
		namespace,
		aclContent,
		provider,
		append(deps, namespace, cert),
	)
	if err != nil {
		return nil, err
	}

	return []pulumi.Resource{namespace, cert, chart}, nil
}

// deployValkeyClusterHelmChart deploys a Valkey cluster using the bitnami Helm chart.
func deployValkeyClusterHelmChart(
	ctx *pulumi.Context,
	namespace *corev1.Namespace,
	aclContent string,
	provider *kubernetes.Provider,
	deps []pulumi.Resource,
) (pulumi.Resource, error) {
	args := newValkeyClusterHelmChartArgs(namespace, aclContent)

	chart, err := helmv4.NewChart(
		ctx,
		chartName,
		args,
		pulumi.Provider(provider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}
	return chart, nil
}

// newValkeyClusterHelmChartArgs constructs the Helm chart values needed to deploy Valkey to k8s.
func newValkeyClusterHelmChartArgs(
	namespace *corev1.Namespace,
	aclContent string,
) *helmv4.ChartArgs {
	chartArgs := &helmv4.ChartArgs{
		Chart:     pulumi.String(chartRepo),
		Namespace: namespace.Metadata.Name(),
		Version:   pulumi.String(chartVersion),
		Values: pulumi.Map{
			"architecture": pulumi.String("replication"),
			"auth": pulumi.Map{
				"enabled":  pulumi.Bool(true),
				"password": pulumi.String("password"),
			},
			"commonConfiguration": pulumi.String(aclContent),
			"sentinel": pulumi.Map{
				"enabled": pulumi.Bool(true),
				"resources": pulumi.Map{
					"limits": pulumi.Map{
						"cpu":    pulumi.String("600m"),
						"memory": pulumi.String("750Mi"),
					},
					"requests": pulumi.Map{
						"cpu":    pulumi.String("500m"),
						"memory": pulumi.String("500Mi"),
					},
				},
			},
			"tls": pulumi.Map{
				"enabled":         pulumi.Bool(true),
				"existingSecret":  pulumi.String(clusterCertificateSecretName),
				"certFilename":    pulumi.String("tls.crt"),
				"certKeyFilename": pulumi.String("tls.key"),
				"certCAFilename":  pulumi.String("ca.crt"),
			},
		},
	}
	return chartArgs
}
