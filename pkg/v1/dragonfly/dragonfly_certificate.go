package dragonfly

import (
	"fmt"
	"github.com/fjarm/infrastructure/pkg/v1/certmanager"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// newDragonflyCertificateArgs creates a new Certificate issued by cert-manager for DragonflyDB to use.
//
// SEE: https://github.com/dragonflydb/dragonfly/blob/v1.30.3/contrib/charts/dragonfly/templates/certificate.yaml
func newDragonflyCertificateArgs(ns *corev1.Namespace) (*apiextensions.CustomResourceArgs, error) {
	cra := apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("Certificate"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String("dragonfly-certificate"),
			Namespace: ns.Metadata.Name(),
			Labels: pulumi.StringMap{
				"app":                          pulumi.String(helmChartName),
				"app.kubernetes.io/instance":   pulumi.String(helmChartName),
				"app.kubernetes.io/managed-by": pulumi.String("Helm"),
				"app.kubernetes.io/name":       pulumi.String(helmChartName),
				"app.kubernetes.io/version":    pulumi.String(chartVersion),
				"helm.sh/chart":                pulumi.String(fmt.Sprintf("%s-%s", helmChartName, chartVersion)),
			},
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": kubernetes.UntypedArgs{
				"commonName": pulumi.String("dragonfly"),
				"dnsNames": pulumi.StringArray{
					pulumi.String("'*.dragonfly.dragonfly.svc.cluster.local'"),
					pulumi.String("dragonfly.dragonfly.svc.cluster.local"),
					pulumi.String("dragonfly.dragonfly.svc"),
					pulumi.String("dragonfly.dragonfly"),
					pulumi.String("dragonfly"),
					pulumi.String("localhost"),
				},
				"duration": pulumi.String("87600h0m0s"),
				"ipAddresses": pulumi.StringArray{
					pulumi.String("127.0.0.1"),
				},
				"issuerRef": kubernetes.UntypedArgs{
					"kind":  pulumi.String("ClusterIssuer"),
					"name":  pulumi.String(certmanager.InternalClusterIssuerName),
					"group": pulumi.String("cert-manager.io"),
				},
				"secretName": pulumi.String(certificateName),
				"usages": pulumi.StringArray{
					pulumi.String("client auth"),
					pulumi.String("server auth"),
					pulumi.String("signing"),
					pulumi.String("key encipherment"),
				},
			},
		},
	}
	return &cra, nil
}
