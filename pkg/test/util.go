package test

import (
	"fmt"

	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

var (
	ConfigNamespace = "testConfig"
	Location        = "uswest"
	Project         = "testProject"
	Stack           = "testStack"
	Tags            = pulumi.StringMap{
		"key": pulumi.String("value"),
	}

	AppSecGroupName            = "test-appsec-group"
	NetworkSecurityRuleName    = "test-network-rule"
	NetworkSecurityGroupName   = "test-network-group"
	PublicIPAllocationMethod   = "Static"
	PublicIPName               = "test-public-ip"
	PublicIPSKU                = "Standard"
	PublicIPVersion            = "IPv4"
	SubnetName                 = "test-subnet"
	ResourceGroupName          = "test-resource-group"
	VirtualNetworkName         = "test-virtual-network"
	VirtualNetworkAddressSpace = "10.0.0.0/16"

	Config = map[string]string{
		// mock application security group
		fmt.Sprintf("%s:appSecurityGroups", ConfigNamespace): `
[{
	"name": "` + AppSecGroupName + `"
}]`,

		// mock network security rules
		fmt.Sprintf("%s:networkSecurityRules", ConfigNamespace): `
[{
  "access": "Allow",
  "description": "test description",
  "destinationAppSecurityGroups": ["` + AppSecGroupName + `"],
  "destinationPortRanges": ["80"],
  "direction": "Inbound",
  "name": "` + NetworkSecurityRuleName + `",
  "priority": 100,
  "protocol": "Tcp",
  "sourceAddressPrefix": "*",
  "sourcePortRange": "*"
}]`,

		// mock network security groups
		fmt.Sprintf("%s:networkSecurityGroups", ConfigNamespace): `
[{
	"name": "` + NetworkSecurityGroupName + `",
	"securityRules": ["` + NetworkSecurityRuleName + `"]
}]`,

		// mock public IP
		fmt.Sprintf("%s:publicIP", ConfigNamespace): `
[{
	"allocationMethod": "` + PublicIPAllocationMethod + `",
	"ipVersion": "` + PublicIPVersion + `",
	"name": "` + PublicIPName + `",
	"sku": "` + PublicIPSKU + `"
}]`,

		// mock subnet
		fmt.Sprintf("%s:subnets", ConfigNamespace): `
[{
	"name": "` + SubnetName + `",
	"addressPrefix": "10.0.0.0/24",
	"securityGroup": "` + NetworkSecurityGroupName + `"
}]`,

		fmt.Sprintf("%s:resourceGroup", ConfigNamespace): `
{
	"location": "` + Location + `",
	"name": "` + ResourceGroupName + `"
}`,

		// mock virtual network
		fmt.Sprintf("%s:virtualNetworks", ConfigNamespace): `
[{
	"name": "` + VirtualNetworkName + `",
	"cidr": "` + VirtualNetworkAddressSpace + `",
	"subnets": ["` + SubnetName + `"]
}]`,
	}
)

func MockApplicationSecurityGroup(ctx *pulumi.Context) (map[string]*network.ApplicationSecurityGroup, error) {
	appSecGroups := map[string]*network.ApplicationSecurityGroup{}
	appSecGroup, err := network.NewApplicationSecurityGroup(ctx, AppSecGroupName, &network.ApplicationSecurityGroupArgs{
		Location:          pulumi.String(Location),
		Name:              pulumi.String(AppSecGroupName),
		ResourceGroupName: pulumi.String(ResourceGroupName),
	})
	if err != nil {
		return nil, err
	}

	appSecGroups[AppSecGroupName] = appSecGroup
	return appSecGroups, nil
}

func MockResourceGroup(ctx *pulumi.Context) (*core.ResourceGroup, error) {
	return core.NewResourceGroup(ctx, ResourceGroupName, &core.ResourceGroupArgs{
		Location: pulumi.String(Location),
		Name:     pulumi.String(ResourceGroupName),
	})
}

func MockVirtualNetworks(ctx *pulumi.Context) (map[string]*network.VirtualNetwork, error) {

	subnet := &network.VirtualNetworkSubnetArgs{}

	virtualNetworks := map[string]*network.VirtualNetwork{}
	virtualNetwork, err := network.NewVirtualNetwork(ctx, VirtualNetworkName,
		&network.VirtualNetworkArgs{
			AddressSpaces:     pulumi.StringArray{pulumi.String(VirtualNetworkAddressSpace)},
			Location:          pulumi.String(Location),
			Name:              pulumi.String(VirtualNetworkName),
			ResourceGroupName: pulumi.String(ResourceGroupName),
			Subnets:           network.VirtualNetworkSubnetArray{subnet},
		})
	if err != nil {
		return nil, err
	}

	virtualNetworks[VirtualNetworkName] = virtualNetwork
	return virtualNetworks, nil
}
