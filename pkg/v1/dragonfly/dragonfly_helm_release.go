package dragonfly

import (
	"fmt"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	certificateName          = "dragonfly-server-tls"
	chartRepo                = "oci://ghcr.io/dragonflydb/dragonfly/helm/dragonfly"
	chartNamespace           = "dragonfly"
	chartVersion             = "v1.30.3"
	exportDragonflyNamespace = "dragonflyNamespace"
	exportDragonflyStatus    = "dragonflyStatus"
	helmChartName            = "dragonfly"
	healthProbeMountPath     = "/scripts"
	probeFileMode            = 0744
	tlsMountPath             = "/etc/dragonfly/tls"
)

// DeployDragonfly attempts to deploy DragonflyDB using the official Helm chart.
func DeployDragonfly(ctx *pulumi.Context, provider *kubernetes.Provider, deps []pulumi.Resource) error {
	ns, err := newDragonflyNamespace(ctx)
	if err != nil {
		return err
	}

	certArgs, err := newDragonflyCertificateArgs(ns)
	if err != nil {
		return err
	}

	cert, err := apiextensions.NewCustomResource(
		ctx,
		certificateName,
		certArgs,
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return err
	}

	cm, err := newDragonflyProbeScript(ctx, ns)
	if err != nil {
		return err
	}

	chartArgs := newDragonflyHelmChartArgs(ns, cm)
	dragonflyChart, err := helmv4.NewChart(
		ctx,
		helmChartName,
		chartArgs,
		pulumi.Provider(provider),
		pulumi.DependsOn(append(deps, cert)),
	)
	if err != nil {
		return err
	}

	ctx.Export(exportDragonflyNamespace, ns.Metadata.Name())
	ctx.Export(exportDragonflyStatus, dragonflyChart.ToChartOutput())
	ctx.Export(certificateName, cert.OtherFields)
	return nil
}

// newDragonflyHelmChartArgs constructs the Helm chart values needed to deploy DragonflyDB to k8s.
func newDragonflyHelmChartArgs(ns *corev1.Namespace, cm *corev1.ConfigMap) *helmv4.ChartArgs {
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
				pulumi.String("--dbfilename=dragonfly-data-dump"),
				pulumi.String("--snapshot_cron=0 * * * *"), // Snapshot every hour
				// TODO(2025-06-06): Update this to be more secure
				pulumi.String("--requirepass=password"),
				pulumi.String("--tls"),
				//pulumi.String(fmt.Sprintf("--tls_ca_cert_file=%s/ca.crt", tlsMountPath)),
				pulumi.String(fmt.Sprintf("--tls_cert_file=%s/tls.crt", tlsMountPath)),
				pulumi.String(fmt.Sprintf("--tls_key_file=%s/tls.key", tlsMountPath)),
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
			"extraVolumes": pulumi.MapArray{
				pulumi.Map{
					"name": pulumi.String("tls"),
					"secret": pulumi.Map{
						"secretName": pulumi.String(certificateName),
					},
				},
				pulumi.Map{
					"name": pulumi.String("healthcheck"),
					"configMap": pulumi.Map{
						"name":        cm.Metadata.Name(),
						"defaultMode": pulumi.Int(probeFileMode),
					},
				},
			},
			"extraVolumeMounts": pulumi.MapArray{
				pulumi.Map{
					"name":      pulumi.String("tls"),
					"mountPath": pulumi.String(tlsMountPath),
				},
				pulumi.Map{
					"name":      pulumi.String("healthcheck"),
					"mountPath": pulumi.String(healthProbeMountPath),
				},
			},
			"podAnnotations": pulumi.StringMap{
				"config.kubernetes.io/depends-on": pulumi.String(
					fmt.Sprintf("/namespaces/%s/Secret/%s", chartNamespace, certificateName),
				),
			},
			"replicaCount": pulumi.Int(1),
			"resources": pulumi.Map{
				"limits": pulumi.Map{
					"memory": pulumi.String("2Gi"),
				},
			},
			"probes": pulumi.Map{
				"livenessProbe": pulumi.Map{
					"exec": pulumi.Map{
						"command": pulumi.StringArray{
							pulumi.String("/bin/sh"),
							pulumi.String(fmt.Sprintf("%s/%s", healthProbeMountPath, "custom-healthcheck.sh")),
						},
					},
				},
				"readinessProbe": pulumi.Map{
					"exec": pulumi.Map{
						"command": pulumi.StringArray{
							pulumi.String("/bin/sh"),
							pulumi.String(fmt.Sprintf("%s/%s", healthProbeMountPath, "custom-healthcheck.sh")),
						},
					},
				},
			},
		},
	}
	return chartArgs
}

// newDragonflyProbeScript creates a new script in the Dragonfly DB Pod that can be used to conduct a health check on an
// instance with TLS enabled.
func newDragonflyProbeScript(ctx *pulumi.Context, ns *corev1.Namespace) (*corev1.ConfigMap, error) {
	configMapArgs := &corev1.ConfigMapArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("dragonfly-probe"),
			Namespace: ns.Metadata.Name(),
		},
		Data: pulumi.StringMap{
			"custom-healthcheck.sh": pulumi.String(healthProbe),
		},
	}
	configMap, err := corev1.NewConfigMap(
		ctx,
		fmt.Sprintf("dragonfly-probe-%s", chartNamespace),
		configMapArgs,
	)
	if err != nil {
		return nil, err
	}
	return configMap, nil
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
