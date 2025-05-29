package redis

import (
	"fmt"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strings"
)

const (
	chartName                    = "redis"
	chartRepo                    = "oci://registry-1.docker.io/bitnamicharts"
	chartNamespace               = "redis"
	chartVersion                 = "21.1.6"
	exportRedisNamespace         = "redisNamespace"
	exportRedisStatus            = "redisStatus"
	helmReleaseName              = "redis"
	k8sProviderLogicalNamePrefix = "kubernetes"
)

// DeployRedis attempts to deploy Redis using the Bitnami Helm chart.
func DeployRedis(ctx *pulumi.Context) error {
	k8sProvider, err := kubernetes.NewProvider(ctx, k8sProviderLogicalNamePrefix, nil)
	if err != nil {
		return err
	}

	releaseArgs := NewRedisHelmReleaseArgs()
	redisRelease, err := helmv3.NewRelease(ctx, helmReleaseName, releaseArgs, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	ctx.Export(exportRedisNamespace, redisRelease.Namespace)
	ctx.Export(exportRedisStatus, redisRelease.Status)
	return nil
}

// NewRedisHelmReleaseArgs constructs the Helm chart values needed to deploy Redis to k8s.
func NewRedisHelmReleaseArgs() *helmv3.ReleaseArgs {
	// Remove the trailing slash from the Helm repo URL if it contains one
	trimmedRepo := strings.TrimSuffix(chartRepo, "/")
	chart := fmt.Sprintf("%s/%s", trimmedRepo, chartName)

	releaseArgs := &helmv3.ReleaseArgs{
		Chart:           pulumi.String(chart),
		Namespace:       pulumi.String(chartNamespace),
		Version:         pulumi.String(chartVersion),
		Atomic:          pulumi.Bool(true),
		CreateNamespace: pulumi.Bool(true),
		DisableCRDHooks: pulumi.Bool(false),
		Timeout:         pulumi.Int(120),
		Values: pulumi.Map{
			// SEE: https://github.com/bitnami/charts/issues/14327#issuecomment-1732543394
			"volumePermissions": pulumi.Map{
				"enabled": pulumi.Bool(true),
			},
		},
	}
	return releaseArgs
}

// VerifyCertManagerDeployed references the certmanager Pulumi Stack to verify that the certmanager Helm chart has
// been deployed. Without this chart, we cannot request TLS certificates.
func VerifyCertManagerDeployed() error {
	return ErrUnimplemented
}
