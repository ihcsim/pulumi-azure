package bastion

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ihcsim/pulumi-azure/v2/pkg/mock"
	"github.com/ihcsim/pulumi-azure/v2/pkg/test"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func TestReconcile(t *testing.T) {
	if err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, test.ConfigNamespace)

		resourceGroup, err := test.MockResourceGroup(ctx)
		if err != nil {
			return err
		}

		publicIPs, err := test.MockPublicIPs(ctx)
		if err != nil {
			return err
		}

		virtualNetworks, err := test.MockVirtualNetworks(ctx)
		if err != nil {
			return err
		}

		bastions, err := Reconcile(ctx, cfg, publicIPs, resourceGroup, virtualNetworks, test.Tags)
		if err != nil {
			return err
		}

		if len(bastions) != 1 {
			return fmt.Errorf("expected 1 bastion host only: %s", test.BastionName)
		}

		var (
			bastion = bastions[0]
			wg      = sync.WaitGroup{}
		)

		wg.Add(1)
		pulumi.All(bastion.Location, bastion.Name, bastion.ResourceGroupName).ApplyT(func(actuals []interface{}) error {
			defer wg.Done()

			if actual := actuals[0].(string); actual != test.Location {
				t.Errorf("locations mismatch. expected: %s, actual: %s", test.Location, actual)
			}

			if actual := actuals[1].(string); actual != test.BastionName {
				t.Errorf("bastion hosts mismatch. expected: %s, actual: %s", test.BastionName, actual)
			}

			if actual := actuals[2].(string); actual != test.ResourceGroupName {
				t.Errorf("resource group names mismatch. expected: %s, actual: %s", test.ResourceGroupName, actual)
			}
			return nil
		})

		wg.Wait()
		return nil
	}, mock.WithCustomMocks(test.Project, test.Stack, test.Config, mock.Mocks(0))); err != nil {
		t.Error(err)
	}
}
