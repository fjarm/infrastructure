package dragonfly

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const clusterACLSecretName = "dragonfly-cluster-acls"

// deployDragonflyClusterACLSecret constructs and deploys a secret containing the Dragonfly ACL.
func deployDragonflyClusterACLSecret(
	ctx *pulumi.Context,
	namespace *corev1.Namespace,
	provider *kubernetes.Provider,
	deps []pulumi.Resource,
) (*corev1.Secret, error) {
	aclContent, err := parseDragonflyClusterUserToTemplate(dragonflyClusterUser{
		Username:        "test",
		Password:        "password",
		EnabledCommands: []string{"+@ALL"},
	})
	if err != nil {
		return nil, err
	}

	aclSecretArgs := newDragonflyClusterACLSecretArgs(namespace, aclContent)
	aclSecret, err := corev1.NewSecret(
		ctx,
		clusterACLSecretName,
		aclSecretArgs,
		pulumi.Provider(provider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}
	return aclSecret, nil
}

// newDragonflyClusterACLSecretArgs sets up a secret containing the Dragonfly ACL file contents.
func newDragonflyClusterACLSecretArgs(
	namespace *corev1.Namespace,
	aclContent string,
) *corev1.SecretArgs {
	return &corev1.SecretArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Secret"),
		Type:       pulumi.String("Opaque"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(clusterACLSecretName),
			Namespace: namespace.Metadata.Name(),
			Labels: pulumi.StringMap{
				"app":                        pulumi.String(clusterAppLabel),
				"app.kubernetes.io/instance": pulumi.String(clusterInstanceLabel),
				"app.kubernetes.io/name":     pulumi.String(clusterNameLabel),
				"app.kubernetes.io/part-of":  pulumi.String(operatorAppLabel),
			},
		},
		StringData: pulumi.StringMap{
			clusterACLSecretName: pulumi.String(aclContent),
		},
	}
}
