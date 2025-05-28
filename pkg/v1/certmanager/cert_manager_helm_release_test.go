package certmanager

import (
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"testing"
)

type mocks struct {
	pulumi.MockResourceMonitor
}

func (m mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func (m mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	// Create a default ID for all resources
	id := args.Name + "_id"

	// For the Kubernetes provider specifically, return a canned output
	if args.TypeToken == "pulumi:providers:kubernetes" {
		return id, resource.NewPropertyMapFromMap(map[string]any{
			"kubeconfig": "mock-kubeconfig-data",
			"context":    "mock-context",
		}), nil
	}

	// For Helm releases, return a mock status output
	if args.TypeToken == "kubernetes:helm.sh/v3:Release" {
		return id, resource.NewPropertyMapFromMap(map[string]any{
			"status": helmv3.ReleaseStatus{
				AppVersion: pulumi.StringRef("mock-app-version"),
				Chart:      pulumi.StringRef("mock-chart"),
				Name:       pulumi.StringRef(args.Name),
				Namespace:  pulumi.StringRef("mock-namespace"),
				Version:    pulumi.StringRef("mock-version"),
				Revision:   pulumi.IntRef(1),
			},
			"name": args.Name,
		}), nil
	}

	// Default behavior for other resources
	return id, args.Inputs, nil
}

func TestCertManagerHelmRelease(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		err := DeployCertManager(ctx)
		return err
	}, pulumi.WithMocks("project", "stack", mocks{}))

	if err != nil {
		t.Errorf("failed to create cert-manager Helm Release resource: %v", err)
	}
}
