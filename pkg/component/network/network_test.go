package network

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ihcsim/pulumi-azure/v2/pkg/mock"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

var (
	testCfgNamespace = "testConfig"
	testLocation     = "uswest"
	testProject      = "testProject"
	testStack        = "testStack"
)

func TestReconcile(t *testing.T) {
	if err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		var (
			cfg  = config.New(ctx, testCfgNamespace)
			tags = pulumi.StringMap{
				"key": pulumi.String("value"),
			}
		)

		resourceGroup, err := core.NewResourceGroup(ctx, testResourceGroupName, &core.ResourceGroupArgs{
			Location: pulumi.String(testLocation),
			Name:     pulumi.String(testResourceGroupName),
		})
		if err != nil {
			return err
		}

		appSecGroup, err := network.NewApplicationSecurityGroup(ctx, testAppSecGroupName, &network.ApplicationSecurityGroupArgs{
			Location:          pulumi.String("uswest"),
			Name:              pulumi.String(testAppSecGroupName),
			ResourceGroupName: resourceGroup.Name,
		})
		if err != nil {
			return err
		}
		appSecGroups := map[string]*network.ApplicationSecurityGroup{
			testAppSecGroupName: appSecGroup,
		}

		virtualNetworks, err := Reconcile(ctx, cfg, appSecGroups, resourceGroup, tags)
		if err != nil {
			return err
		}

		virtualNetwork, exists := virtualNetworks[testVirtualNetworkName]
		if !exists {
			t.Errorf("missing virtual network: %s", testVirtualNetworkName)
		}

		var wg sync.WaitGroup
		wg.Add(1)
		pulumi.All(virtualNetwork.Location, virtualNetwork.ResourceGroupName).ApplyT(func(actuals []interface{}) error {
			defer wg.Done()

			if actual := actuals[0].(string); actual != testLocation {
				t.Errorf("locations mismatch. expected: %s, actual: %s", testLocation, actual)
			}

			if actual := actuals[1].(string); actual != testResourceGroupName {
				t.Errorf("resource group names mismatch. expected: %s, actual: %s", testResourceGroupName, actual)
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
				if actual := actuals[0].(*string); *actual != "10.0.0.0/24" {
					t.Errorf("address prefix mismatch. expected: 10.0.0.0/24, actual: %s", *actual)
				}

				if actual := actuals[1].(string); actual != testNetworkSecurityGroupName {
					t.Errorf("security group name mismatch. expected: %s, actual: %s", testNetworkSecurityGroupName, actual)
				}

				return nil
			})

			return nil
		})

		wg.Wait()
		return nil
	}, mock.WithCustomMocks(testProject, testStack, testConfig, mock.Mocks(0))); err != nil {
		t.Error(err)
	}
}

var (
	testAppSecGroupName          = "test-appsec-group"
	testNetworkSecurityRuleName  = "test-network-rule"
	testNetworkSecurityGroupName = "test-network-group"
	testSubnetName               = "test-subnet"
	testResourceGroupName        = "test-resource-group"
	testVirtualNetworkName       = "test-virtual-network"

	testConfig = map[string]string{
		fmt.Sprintf("%s:networkSecurityRules", testCfgNamespace): `
[{
  "access": "Allow",
  "description": "test description",
  "destinationAppSecurityGroups": ["` + testAppSecGroupName + `"],
  "destinationPortRanges": ["80"],
  "direction": "Inbound",
  "name": "` + testNetworkSecurityRuleName + `",
  "priority": 100,
  "protocol": "Tcp",
  "sourceAddressPrefix": "*",
  "sourcePortRange": "*"
}]`,
		fmt.Sprintf("%s:networkSecurityGroups", testCfgNamespace): `
[{
	"name": "` + testNetworkSecurityGroupName + `",
	"securityRules": ["` + testNetworkSecurityRuleName + `"]
}]`,
		fmt.Sprintf("%s:subnets", testCfgNamespace): `
[{
	"name": "` + testSubnetName + `",
	"addressPrefix": "10.0.0.0/24",
	"securityGroup": "` + testNetworkSecurityGroupName + `"
}]`,
		fmt.Sprintf("%s:virtualNetworks", testCfgNamespace): `
[{
	"name": "` + testVirtualNetworkName + `",
	"cidr": "10.0.0.0/16",
	"subnets": ["` + testSubnetName + `"]
}]`,
	}
)
