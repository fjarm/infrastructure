package release

import (
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
		return id, resource.NewPropertyMapFromMap(map[string]interface{}{
			"kubeconfig": "mock-kubeconfig-data",
			"context":    "mock-context",
		}), nil
	}

	// For Helm releases, return a mock status output
	if args.TypeToken == "kubernetes:helm.sh/v3:Release" {
		return id, resource.NewPropertyMapFromMap(map[string]interface{}{
			"status": "deployed",
			"name":   args.Name,
		}), nil
	}

	// Default behavior for other resources
	return id, args.Inputs, nil
	//return args.Name + "_id", args.Inputs, nil
}

func TestCertManagerHelmRelease(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		_, err := NewCertManagerHelmReleaseArgs()
		if err != nil {
			t.Errorf("NewCertManagerHelmReleaseArgs failed: %v", err)
		}
		return nil
	}, pulumi.WithMocks("project", "stack", mocks{}))

	if err != nil {
		t.Errorf("failed to create cert-manager Helm Release resource: %v", err)
	}
}
