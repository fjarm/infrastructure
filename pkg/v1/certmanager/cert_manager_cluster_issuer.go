package certmanager

import (
	"github.com/fjarm/infrastructure/pkg/v1/common"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	certManagerCACertName     = "cert-manager-ca-cert"
	exportedSelfSignedCertKey = "selfSignedCertKey"
	exportedSelfSignedCertPem = "selfSignedCertPem"
	InternalClusterIssuerName = "internal-cluster-issuer"
	privateKeyName            = "privateKey"
	selfSignedCertName        = "selfSignedCert"
)

// deployCertManagerInternalClusterIssuer deploys an internal ClusterIssuer resource. If the Kubernetes cluster is
// local - i.e. Kind or Minikube - then it simply uses a self-signed certificate generated from Pulumi's provider.
func deployCertManagerInternalClusterIssuer(
	ctx *pulumi.Context,
	k8sProvider *kubernetes.Provider,
	deps []pulumi.Resource,
	kind bool,
) (*apiextensions.CustomResource, error) {
	if !kind {
		return nil, common.ErrUnimplemented
	}

	key, cert, err := newPulumiRootCACertificate(ctx, kind)
	if err != nil {
		return nil, err
	}

	secret, err := deploySecretFromCACertificate(ctx, k8sProvider, key, cert, []pulumi.Resource{key, cert})
	if err != nil {
		return nil, err
	}

	cra := newCertManagerInternalClusterIssuerArgs()
	clusterIssuer, err := apiextensions.NewCustomResource(
		ctx,
		InternalClusterIssuerName,
		cra,
		pulumi.DependsOn(append(deps, secret)),
	)
	if err != nil {
		return nil, err
	}
	return clusterIssuer, nil
}

// deploySecretFromCACertificate deploys a Secret to the k8s cluster that can be used to bootstrap a ClusterIssuer.
func deploySecretFromCACertificate(
	ctx *pulumi.Context,
	k8sProvider *kubernetes.Provider,
	key *tls.PrivateKey,
	cert *tls.SelfSignedCert,
	deps []pulumi.Resource,
) (*corev1.Secret, error) {
	secretArgs := corev1.SecretArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Secret"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(certManagerCACertName),
			Namespace: pulumi.String(chartNamespace),
		},
		Type: pulumi.String("kubernetes.io/tls"),
		StringData: pulumi.StringMap{
			"tls.key": key.PrivateKeyPem,
			"tls.crt": cert.CertPem,
		},
	}
	secret, err := corev1.NewSecret(
		ctx,
		certManagerCACertName,
		&secretArgs,
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// newCertManagerInternalClusterIssuerArgs returns a pointer to apiextensions.CustomResourceArgs that sets up a
// ClusterIssuer with a reference to self-signed certificates created by Pulumi.
func newCertManagerInternalClusterIssuerArgs() *apiextensions.CustomResourceArgs {
	cra := apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("ClusterIssuer"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(InternalClusterIssuerName),
			Namespace: pulumi.String(chartNamespace),
			Annotations: pulumi.StringMap{
				"config.kubernetes.io/depends-on": pulumi.String(
					"/namespaces/cert-manager/Deployment/cert-manager-webhook",
				),
			},
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"ca": map[string]any{
					"secretName": certManagerCACertName,
				},
			},
		},
	}
	return &cra
}

// newPulumiRootCACertificate will use Pulumi do create a new root CA certificate to be used to stand up a secret used by a CA
// ClusterIssuer. If deploying locally to Kind or Minikube, it'll be a self-signed certificate. Otherwise, it'll be a
// certificate from the Infisical PKI.
func newPulumiRootCACertificate(ctx *pulumi.Context, kind bool) (*tls.PrivateKey, *tls.SelfSignedCert, error) {
	if kind {
		key, err := tls.NewPrivateKey(ctx, privateKeyName, &tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
		})
		if err != nil {
			return nil, nil, err
		}
		cert, err := tls.NewSelfSignedCert(ctx, selfSignedCertName, &tls.SelfSignedCertArgs{
			PrivateKeyPem: key.PrivateKeyPem,
			AllowedUses: pulumi.StringArray{
				pulumi.String("cert_signing"),
				pulumi.String("client_auth"),
				pulumi.String("digital_signature"),
				pulumi.String("server_auth"),
			},
			IsCaCertificate: pulumi.Bool(true),
			Subject: &tls.SelfSignedCertSubjectArgs{
				Organization: pulumi.String("Fjarm"),
			},
			ValidityPeriodHours: pulumi.Int(807660),
		})
		if err != nil {
			return nil, nil, err
		}

		ctx.Export(exportedSelfSignedCertKey, key.PrivateKeyPem)
		ctx.Export(exportedSelfSignedCertPem, cert.CertPem)

		return key, cert, nil
	}
	return nil, nil, common.ErrUnimplemented
}
