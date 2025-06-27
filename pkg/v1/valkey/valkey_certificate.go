package valkey

import (
	"fmt"
	"github.com/fjarm/infrastructure/pkg/v1/certmanager"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	clusterCertificateName       = "valkey-certificate"
	clusterCertificateSecretName = "valkey-tls-secret"
)

// deployValkeyClusterCertificate deploys a cert-manager created/managed TLS certificate in the `valkey` namespace.
func deployValkeyClusterCertificate(
	ctx *pulumi.Context,
	namespace *corev1.Namespace,
	provider *kubernetes.Provider,
	deps []pulumi.Resource,
) (*apiextensions.CustomResource, error) {
	certArgs, err := newValkeyClusterCertificateArgs(namespace)
	if err != nil {
		return nil, err
	}

	cert, err := apiextensions.NewCustomResource(
		ctx,
		clusterCertificateName,
		certArgs,
		pulumi.Provider(provider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

// newValkeyClusterCertificateArgs creates a new Certificate issued by cert-manager for Valkey to use.
func newValkeyClusterCertificateArgs(ns *corev1.Namespace) (*apiextensions.CustomResourceArgs, error) {
	cra := apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("Certificate"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(clusterCertificateName),
			Namespace: ns.Metadata.Name(),
			Labels: pulumi.StringMap{
				"app":                          pulumi.String(clusterAppLabel),
				"app.kubernetes.io/managed-by": pulumi.String("Helm"),
				"app.kubernetes.io/version":    pulumi.String(chartVersion),
				"helm.sh/chart":                pulumi.String(fmt.Sprintf("%s-%s", chartName, chartVersion)),
			},
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": kubernetes.UntypedArgs{
				"commonName": pulumi.String("valkey"),
				"dnsNames": pulumi.StringArray{
					pulumi.String("'*.valkey.valkey.svc.cluster.local'"),
					pulumi.String("valkey.valkey.svc.cluster.local"),
					pulumi.String("valkey.valkey.svc"),
					pulumi.String("valkey.valkey"),
					pulumi.String("valkey"),
					pulumi.String("'*.valkey-headless.valkey.svc.cluster.local'"),
					pulumi.String("valkey-headless.valkey.svc.cluster.local"),
					pulumi.String("valkey-headless.valkey.svc"),
					pulumi.String("valkey-headless.valkey"),
					pulumi.String("valkey-headless"),
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
				"secretName": pulumi.String(clusterCertificateSecretName),
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
