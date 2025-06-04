package certmanager

import (
	"github.com/fjarm/infrastructure/pkg/v1/common"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	k8sV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	k8sV2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	exportedSelfSignedCertKey = "selfSignedCertKey"
	exportedSelfSignedCertPem = "selfSignedCertPem"
	privateKeyName            = "privateKey"
	selfSignedCertName        = "selfSignedCert"
)

// DeployCertManagerInternalClusterIssuer deploys an internal ClusterIssuer resource. If the Kubernetes cluster is
// local - i.e. Kind or Minikube - then it simply uses a self-signed certificate generated from Pulumi's provider.
func DeployCertManagerInternalClusterIssuer(
	ctx *pulumi.Context,
	k8sProvider *kubernetes.Provider,
	deps []pulumi.Resource,
	kind bool,
) error {
	if !kind {
		return common.ErrUnimplemented
	}

	cert, err := newPulumiRootCACertificate(ctx, kind)
	if err != nil {
		return err
	}

	secret, err := DeploySecretFromCACertificate(ctx, k8sProvider, []pulumi.Resource{cert})
	if err != nil {
		return err
	}

	cra := NewCertManagerInternalClusterIssuerArgs()
	_, err = apiextensions.NewCustomResource(
		ctx,
		"internal-cluster-issuer",
		cra,
		pulumi.DependsOn(append(deps, secret)),
	)
	if err != nil {
		return err
	}
	return nil
}

// DeploySecretFromCACertificate deploys a Secret to the k8s cluster that can be used to bootstrap a ClusterIssuer.
func DeploySecretFromCACertificate(
	ctx *pulumi.Context,
	k8sProvider *kubernetes.Provider,
	deps []pulumi.Resource,
) (*k8sV1.Secret, error) {
	stack, err := pulumi.NewStackReference(ctx, "veganafro/infrastructure/dev", nil)
	if err != nil {
		return nil, err
	}

	certKey := stack.GetStringOutput(pulumi.String(exportedSelfSignedCertKey))
	certPem := stack.GetStringOutput(pulumi.String(exportedSelfSignedCertPem))

	secretArgs := k8sV1.SecretArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Secret"),
		Metadata: &k8sV2.ObjectMetaArgs{
			Name:      pulumi.String("cert-manager-ca-cert"),
			Namespace: pulumi.String(chartNamespace),
		},
		Type: pulumi.String("kubernetes.io/tls"),
		Data: pulumi.StringMap{
			"tls.key": certKey,
			"tls.crt": certPem,
		},
	}
	secret, err := k8sV1.NewSecret(
		ctx,
		"cert-manager-ca-cert",
		&secretArgs,
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// NewCertManagerInternalClusterIssuerArgs returns a pointer to apiextensions.CustomResourceArgs that sets up a
// ClusterIssuer with a reference to self-signed certificates created by Pulumi.
func NewCertManagerInternalClusterIssuerArgs() *apiextensions.CustomResourceArgs {
	cra := apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("ClusterIssuer"),
		Metadata: k8sV2.ObjectMetaArgs{
			Name:      pulumi.String("internal-cluster-issuer"),
			Namespace: pulumi.String(chartNamespace),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"ca": map[string]any{
					"secretName": "cert-manager-ca-cert",
				},
			},
		},
	}
	return &cra
}

// newPulumiRootCACertificate will use Pulumi do create a new root CA certificate to be used to stand up a secret used by a CA
// ClusterIssuer. If deploying locally to Kind or Minikube, it'll be a self-signed certificate. Otherwise, it'll be a
// certificate from the Infisical PKI.
func newPulumiRootCACertificate(ctx *pulumi.Context, kind bool) (*tls.SelfSignedCert, error) {
	if kind {
		key, err := tls.NewPrivateKey(ctx, privateKeyName, &tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
		})
		if err != nil {
			return nil, err
		}
		cert, err := tls.NewSelfSignedCert(ctx, selfSignedCertName, &tls.SelfSignedCertArgs{
			PrivateKeyPem: key.PrivateKeyPem,
			AllowedUses: pulumi.StringArray{
				pulumi.String("cert_signing"),
				pulumi.String("client_auth"),
				pulumi.String("digital_signature"),
				pulumi.String("server_auth"),
			},
			DnsNames: pulumi.StringArray{
				pulumi.String("fjarm.xyz"),
			},
			IsCaCertificate:   pulumi.Bool(true),
			SetAuthorityKeyId: pulumi.Bool(true),
			SetSubjectKeyId:   pulumi.Bool(true),
			Subject: &tls.SelfSignedCertSubjectArgs{
				Organization: pulumi.String("Fjarm"),
			},
			ValidityPeriodHours: pulumi.Int(807660),
		})
		if err != nil {
			return nil, err
		}

		ctx.Export(exportedSelfSignedCertKey, cert.PrivateKeyPem)
		ctx.Export(exportedSelfSignedCertPem, cert.CertPem)

		return cert, nil
	}
	return nil, common.ErrUnimplemented
}
