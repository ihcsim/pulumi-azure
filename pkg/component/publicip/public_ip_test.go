package publicip

import (
	"sync"
	"testing"

	"github.com/ihcsim/pulumi-azure/v2/pkg/mock"
	"github.com/ihcsim/pulumi-azure/v2/pkg/test"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func TestReconcile(t *testing.T) {
	if err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		var (
			cfg  = config.New(ctx, test.ConfigNamespace)
			tags = test.Tags
		)

		resourceGroup, err := test.MockResourceGroup(ctx)
		if err != nil {
			return err
		}

		publicIPs, err := Reconcile(ctx, cfg, resourceGroup, tags)
		if err != nil {
			return err
		}

		publicIP, exists := publicIPs[test.PublicIPName]
		if !exists {
			t.Fatalf("missing public IP: %s", test.PublicIPName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		pulumi.All(publicIP.AllocationMethod, publicIP.IpVersion, publicIP.Name, publicIP.Sku).ApplyT(func(actuals []interface{}) error {
			defer wg.Done()

			if actual := actuals[0].(string); actual != test.PublicIPAllocationMethod {
				t.Errorf("allocation method mismatch. expected: %s, actual: %s", test.PublicIPAllocationMethod, actual)
			}

			if actual := actuals[1].(string); actual != test.PublicIPVersion {
				t.Errorf("IP version mismatch. expected: %s, actual: %s", test.PublicIPVersion, actual)
			}

			if actual := actuals[2].(string); actual != test.PublicIPSKU {
				t.Errorf("public IP SKU mismatch. expected: %s, actual: %s", test.PublicIPSKU, actual)
			}

			wg.Wait()
			return nil
		})

		return nil
	}, mock.WithCustomMocks(test.Project, test.Stack, test.Config, mock.Mocks(0))); err != nil {
		t.Error(err)
	}
}
