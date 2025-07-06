package dragonfly

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	clusterNamespace = "dragonfly-cluster"
)

// deployDragonflyClusterNamespace creates the `dragonfly-cluster` namespace.
func deployDragonflyClusterNamespace(
	ctx *pulumi.Context,
	provider *kubernetes.Provider,
) (*corev1.Namespace, error) {
	args := newDragonflyClusterNamespaceArgs()
	ns, err := corev1.NewNamespace(
		ctx,
		clusterNamespace,
		args,
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// newDragonflyClusterNamespaceArgs returns the corev1.NamespaceArgs used to create the `dragonfly-cluster` namespace.
func newDragonflyClusterNamespaceArgs() *corev1.NamespaceArgs {
	return &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(clusterNamespace),
			Labels: pulumi.StringMap{
				"app": pulumi.String(clusterAppLabel),
			},
		},
	}
}
