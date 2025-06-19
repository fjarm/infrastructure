package dragonfly

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	clusterAppLabel        = "dragonfly"
	clusterInstanceLabel   = "dragonfly-instance"
	clusterName            = "dragonfly-cluster"
	clusterNameLabel       = "dragonfly-cluster"
	dragonflyCRDAPIVersion = "dragonflydb.io/v1alpha1"
	dragonflyCRDKind       = "Dragonfly"
)

// deployDragonflyCluster deploys a DragonflyDB cluster using the Dragonfly Operator's CRD.
func deployDragonflyCluster(
	ctx *pulumi.Context,
	namespace *corev1.Namespace,
	provider *kubernetes.Provider,
	deps []pulumi.Resource,
) (pulumi.Resource, error) {
	clusterArgs := newDragonflyClusterArgs(namespace)
	cluster, err := apiextensions.NewCustomResource(
		ctx,
		clusterName,
		clusterArgs,
		pulumi.Provider(provider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// newDragonflyClusterArgs returns the Dragonfly cluster spec that will be converted to YAML and applied to the cluster.
func newDragonflyClusterArgs(
	namespace *corev1.Namespace,
) *apiextensions.CustomResourceArgs {
	return &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String(dragonflyCRDAPIVersion),
		Kind:       pulumi.String(dragonflyCRDKind),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(clusterName),
			Namespace: namespace.Metadata.Name(),
			Labels: pulumi.StringMap{
				"app":                        pulumi.String(clusterAppLabel),
				"app.kubernetes.io/instance": pulumi.String(clusterInstanceLabel),
				"app.kubernetes.io/name":     pulumi.String(clusterNameLabel),
				"app.kubernetes.io/part-of":  pulumi.String(operatorAppLabel),
			},
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": kubernetes.UntypedArgs{
				"replicas": pulumi.Int(3), // 1 primary and 2 replicas
				"image":    pulumi.String("docker.dragonflydb.io/dragonflydb/dragonfly:v1.30.3"),
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
				"args": pulumi.StringArray{
					pulumi.String("--bind=127.0.0.1"), // IPv4 loopback interface
					// SEE: https://github.com/dragonflydb/dragonfly-operator/blob/c70f578b5c64372ca41718be830e0629f2f144ef/internal/resources/resources.go#L114-L115
					// SEE: https://github.com/dragonflydb/dragonfly-operator/blob/3425e0b0623f785084f0cf1463bc58ca379bf408/internal/resources/const.go#L101
					// Dragonfly hard-codes HEALTHCHECK_PORT and --admin_port to 9999 and overrides don't work.
					//pulumi.String("--admin_port=9999"),
					pulumi.String("--dir=/data"),
					pulumi.String("--dbfilename=dragonfly-dump"), // No {timestamp} macro to not create multiple files that consume space on the volume.
					pulumi.String("--snapshot_cron=0 * * * *"),   // Snapshot at minute 0 of every hour
				},
			},
		},
	}
}
