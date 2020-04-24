package network

import (
	"sync"
	"testing"

	"github.com/ihcsim/pulumi-azure/v2/pkg/mock"
	"github.com/ihcsim/pulumi-azure/v2/pkg/test"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
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

		appSecGroup, err := network.NewApplicationSecurityGroup(ctx, test.AppSecGroupName, &network.ApplicationSecurityGroupArgs{
			Location:          resourceGroup.Location,
			Name:              pulumi.String(test.AppSecGroupName),
			ResourceGroupName: resourceGroup.Name,
		})
		if err != nil {
			return err
		}
		appSecGroups := map[string]*network.ApplicationSecurityGroup{
			test.AppSecGroupName: appSecGroup,
		}

		virtualNetworks, err := Reconcile(ctx, cfg, appSecGroups, resourceGroup, tags)
		if err != nil {
			return err
		}

		virtualNetwork, exists := virtualNetworks[test.VirtualNetworkName]
		if !exists {
			t.Errorf("missing virtual network: %s", test.VirtualNetworkName)
		}

		var wg sync.WaitGroup
		wg.Add(1)
		pulumi.All(virtualNetwork.Location, virtualNetwork.ResourceGroupName).ApplyT(func(actuals []interface{}) error {
			defer wg.Done()

			if actual := actuals[0].(string); actual != test.Location {
				t.Errorf("locations mismatch. expected: %s, actual: %s", test.Location, actual)
			}

			if actual := actuals[1].(string); actual != test.ResourceGroupName {
				t.Errorf("resource group names mismatch. expected: %s, actual: %s", test.ResourceGroupName, actual)
			}

			return nil
		})

		wg.Add(1)
		virtualNetwork.AddressSpaces.ApplyT(func(addressSpaces []string) error {
			defer wg.Done()

			if len(addressSpaces) != 1 {
				t.Error("expected address spaces list to have a length of 1")
			}

			if actual := addressSpaces[0]; actual != "10.0.0.0/16" {
				t.Errorf("address spaces mismatch. expected: 10.0.0.0/16, actual: %s", actual)
			}

			return nil
		})

		wg.Add(1)
		virtualNetwork.Subnets.ApplyT(func(subnets []network.VirtualNetworkSubnet) error {
			defer wg.Done()

			if len(subnets) != 1 {
				t.Error("expected number of subnets to be 1")
			}

			subnet := subnets[0]
			pulumi.All(subnet.AddressPrefix, subnet.SecurityGroup).ApplyT(func(actuals []interface{}) error {
				if actual := actuals[0].(string); actual != "10.0.0.0/24" {
					t.Errorf("address prefix mismatch. expected: 10.0.0.0/24, actual: %s", actual)
				}

				if actual := actuals[1].(*string); *actual != test.NetworkSecurityGroupName {
					t.Errorf("security group name mismatch. expected: %s, actual: %s", test.NetworkSecurityGroupName, *actual)
				}

				return nil
			})

			return nil
		})

		wg.Wait()
		return nil
	}, mock.WithCustomMocks(test.Project, test.Stack, test.Config, mock.Mocks(0))); err != nil {
		t.Error(err)
	}
}
