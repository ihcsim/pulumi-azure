package resourcegroup

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

		resourceGroup, err := Reconcile(ctx, cfg, tags)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		wg.Add(1)

		pulumi.All(resourceGroup.Location, resourceGroup.Name).ApplyT(func(actuals []interface{}) error {
			defer wg.Done()

			if actual := actuals[0].(string); actual != test.Location {
				t.Errorf("locations mismatch. expected: %s, actual: %s", test.Location, actual)
			}

			if actual := actuals[1].(string); actual != test.ResourceGroupName {
				t.Errorf("names mismatch. expected: %s, actual: %s", test.ResourceGroupName, actual)
			}

			wg.Wait()
			return nil
		})

		return nil
	}, mock.WithCustomMocks(test.Project, test.Stack, test.Config, mock.Mocks(0))); err != nil {
		t.Error(err)
	}
}
