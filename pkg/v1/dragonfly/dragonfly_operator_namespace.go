package dragonfly

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	operatorNamespace = "dragonfly-operator"
)

// newDragonflyOperatorNamespace creates the `dragonfly-operator` namespace.
func newDragonflyOperatorNamespace(
	ctx *pulumi.Context,
	provider *kubernetes.Provider,
) (*corev1.Namespace, error) {
	args := newDragonflyOperatorNamespaceArgs()
	ns, err := corev1.NewNamespace(
		ctx,
		operatorNamespace,
		args,
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// newDragonflyOperatorNamespaceArgs returns the corev1.NamespaceArgs used to create the `dragonfly-operator` namespace.
func newDragonflyOperatorNamespaceArgs() *corev1.NamespaceArgs {
	return &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(operatorNamespace),
			Labels: pulumi.StringMap{
				"app": pulumi.String(operatorAppLabel),
			},
		},
	}
}
