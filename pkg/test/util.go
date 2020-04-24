package test

import (
	"fmt"

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

	AppSecGroupName          = "test-appsec-group"
	NetworkSecurityRuleName  = "test-network-rule"
	NetworkSecurityGroupName = "test-network-group"
	PublicIPAllocationMethod = "Static"
	PublicIPName             = "test-public-ip"
	PublicIPSKU              = "Standard"
	PublicIPVersion          = "IPv4"
	SubnetName               = "test-subnet"
	ResourceGroupName        = "test-resource-group"
	VirtualNetworkName       = "test-virtual-network"

	Config = map[string]string{
		fmt.Sprintf("%s:appSecurityGroups", ConfigNamespace): `
[{
	"name": "` + AppSecGroupName + `"
}]`,
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

		// mock virtual network
		fmt.Sprintf("%s:virtualNetworks", ConfigNamespace): `
[{
	"name": "` + VirtualNetworkName + `",
	"cidr": "10.0.0.0/16",
	"subnets": ["` + SubnetName + `"]
}]`,
	}
)
