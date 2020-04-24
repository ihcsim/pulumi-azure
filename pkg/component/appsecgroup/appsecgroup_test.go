package appsecgroup

import (
	"sync"
	"testing"

	"github.com/ihcsim/pulumi-azure/v2/pkg/mock"
	"github.com/ihcsim/pulumi-azure/v2/pkg/test"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func TestReconcile(t *testing.T) {
	if err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		var (
			cfg  = config.New(ctx, test.ConfigNamespace)
			tags = test.Tags
		)

		resourceGroup, err := core.NewResourceGroup(ctx, test.ResourceGroupName, &core.ResourceGroupArgs{
			Location: pulumi.String(test.Location),
			Name:     pulumi.String(test.ResourceGroupName),
		})
		if err != nil {
			return err
		}
		appSecGroups, err := Reconcile(ctx, cfg, resourceGroup, tags)
		if err != nil {
			return err
		}

		appSecGroup, exists := appSecGroups[test.AppSecGroupName]
		if !exists {
			t.Fatalf("missing application security group: %s", test.AppSecGroupName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		pulumi.All(appSecGroup.Location, appSecGroup.Name, appSecGroup.ResourceGroupName).ApplyT(func(actuals []interface{}) error {
			defer wg.Done()

			if actual := actuals[0].(string); actual != test.Location {
				t.Errorf("locations mismatch. expected: %s, actual: %s", test.Location, actual)
			}

			if actual := actuals[1].(string); actual != test.AppSecGroupName {
				t.Errorf("names mismatch. expected: %s, actual: %s", test.AppSecGroupName, actual)
			}

			if actual := actuals[2].(string); actual != test.ResourceGroupName {
				t.Errorf("resource groups mismatch. expected: %s, actual: %s", test.ResourceGroupName, actual)
			}

			wg.Wait()
			return nil
		})

		return nil
	}, mock.WithCustomMocks(test.Project, test.Stack, test.Config, mock.Mocks(0))); err != nil {
		t.Error(err)
	}
}
