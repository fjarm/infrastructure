package valkey

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	clusterNamespace = "valkey"
)

// deployValkeyClusterNamespace creates the `valkey` namespace.
func deployValkeyClusterNamespace(
	ctx *pulumi.Context,
	provider *kubernetes.Provider,
	deps []pulumi.Resource,
) (*corev1.Namespace, error) {
	args := newValkeyClusterNamespaceArgs()
	ns, err := corev1.NewNamespace(
		ctx,
		clusterNamespace,
		args,
		pulumi.Provider(provider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// newValkeyClusterNamespaceArgs returns the corev1.NamespaceArgs used to create the `valkey` namespace.
func newValkeyClusterNamespaceArgs() *corev1.NamespaceArgs {
	return &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(clusterNamespace),
			Labels: pulumi.StringMap{
				"app": pulumi.String(clusterAppLabel),
			},
		},
	}
}
