package main

import (
	"fmt"

	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

const (
	owner  = "isim-dev"
	region = pulumi.String("WestUS")
)

var (
	vnetMeta = struct {
		cidr pulumi.StringArray
	}{
		cidr: pulumi.StringArray{pulumi.String("10.0.0.0/16")},
	}

	subnetMeta = []struct {
		name          pulumi.String
		addressPrefix pulumi.String
		securityGroup pulumi.String
	}{
		{name: pulumi.String("subnet-00"), addressPrefix: pulumi.String("10.0.10.0/24")},
		{name: pulumi.String("subnet-01"), addressPrefix: pulumi.String("10.0.20.0/24")},
		{name: pulumi.String("subnet-02"), addressPrefix: pulumi.String("10.0.30.0/24")},
	}

	appSecGroupMeta = []struct {
		name string
	}{
		{name: "web-servers"},
		{name: "admin-servers"},
	}

	networkSecGroupMeta = []struct {
		name string
	}{
		{name: "default"},
	}

	networkRules = []struct {
		access                       pulumi.String
		description                  pulumi.String
		destinationPortRanges        pulumi.StringArray
		destinationAppSecurityGroups pulumi.StringArray
		direction                    pulumi.String
		name                         pulumi.String
		networkSecGroups             pulumi.StringArray
		priority                     pulumi.Int
		protocol                     pulumi.String
		sourceAddressPrefix          pulumi.String
		sourcePortRange              pulumi.String
	}{
		{
			access:                       pulumi.String("Allow"),
			description:                  pulumi.String("allow HTTP and HTTPS"),
			destinationPortRanges:        pulumi.StringArray{pulumi.String("80"), pulumi.String("443")},
			destinationAppSecurityGroups: pulumi.StringArray{pulumi.String(appSecGroupMeta[0].name)},
			direction:                    "Inbound",
			name:                         pulumi.String("allow-web-all"),
			networkSecGroups:             pulumi.StringArray{pulumi.String(networkSecGroupMeta[0].name)},
			priority:                     100,
			protocol:                     pulumi.String("Tcp"),
			sourceAddressPrefix:          "AzureLoadBalancer",
			sourcePortRange:              "*",
		},
		{
			access:                       pulumi.String("Allow"),
			description:                  pulumi.String("allow SSH"),
			destinationPortRanges:        pulumi.StringArray{pulumi.String("22")},
			destinationAppSecurityGroups: pulumi.StringArray{pulumi.String(appSecGroupMeta[1].name)},
			direction:                    "Inbound",
			name:                         pulumi.String("allow-ssh-all"),
			networkSecGroups:             pulumi.StringArray{pulumi.String(networkSecGroupMeta[0].name)},
			priority:                     101,
			protocol:                     pulumi.String("Tcp"),
			sourceAddressPrefix:          "*",
			sourcePortRange:              "*",
		},
	}
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		commonTags := pulumi.StringMap{
			"project": pulumi.String(ctx.Project()),
			"stack":   pulumi.String(ctx.Stack()),
		}

		resourceGroup, err := core.NewResourceGroup(ctx, owner,
			&core.ResourceGroupArgs{
				Location: region,
				Tags:     commonTags,
			})
		if err != nil {
			return err
		}

		// a map of appSecgroup name => appSecGroup instance
		appSecGroups := map[pulumi.String]*network.ApplicationSecurityGroup{}
		for _, groupMeta := range appSecGroupMeta {
			sg, err := network.NewApplicationSecurityGroup(ctx, groupMeta.name,
				&network.ApplicationSecurityGroupArgs{
					Location:          resourceGroup.Location,
					Name:              pulumi.String(groupMeta.name),
					ResourceGroupName: resourceGroup.Name,
					Tags:              commonTags,
				})
			if err != nil {
				return err
			}

			appSecGroups[pulumi.String(groupMeta.name)] = sg
		}

		networkSecRules := map[pulumi.String]network.NetworkSecurityGroupSecurityRuleArgs{}
		for _, rule := range networkRules {
			// bind the application security group to the network rules
			// using the `appSecGroups` map created above
			appSecGroupIDs := pulumi.StringArray{}
			for _, appID := range rule.destinationAppSecurityGroups {
				appSecGroupIDs = append(appSecGroupIDs, appSecGroups[appID.(pulumi.String)].ID())
			}

			for _, secgroup := range rule.networkSecGroups {
				networkSecRules[secgroup.(pulumi.String)] = network.NetworkSecurityGroupSecurityRuleArgs{
					Access:                                 rule.access,
					Description:                            rule.description,
					DestinationPortRanges:                  rule.destinationPortRanges,
					DestinationApplicationSecurityGroupIds: appSecGroupIDs,
					Direction:                              rule.direction,
					Name:                                   rule.name,
					Priority:                               rule.priority,
					Protocol:                               rule.protocol,
					SourceAddressPrefix:                    rule.sourceAddressPrefix,
					SourcePortRange:                        rule.sourcePortRange,
				}
			}
		}

		networkSecGroups := []*network.NetworkSecurityGroup{}
		for _, meta := range networkSecGroupMeta {
			secGroup, err := network.NewNetworkSecurityGroup(ctx, meta.name,
				&network.NetworkSecurityGroupArgs{
					Location:          resourceGroup.Location,
					ResourceGroupName: resourceGroup.Name,
					SecurityRules:     networkSecRules[meta.name],
					Tags:              commonTags,
				})
			if err != nil {
				return err
			}
			networkSecGroups = append(networkSecGroups, secGroup)
		}

		subnets := network.VirtualNetworkSubnetArray{}
		for _, meta := range subnetMeta {
			subnets = append(subnets, network.VirtualNetworkSubnetArgs{
				AddressPrefix: meta.addressPrefix,
				Name:          meta.name,
				SecurityGroup: networkSecGroups[0].ID(),
			})
		}
		vnet, err := network.NewVirtualNetwork(ctx, owner,
			&network.VirtualNetworkArgs{
				AddressSpaces:     vnetMeta.cidr,
				Location:          resourceGroup.Location,
				ResourceGroupName: resourceGroup.Name,
				Tags:              commonTags,
				Subnets:           subnets,
			})
		if err != nil {
			return err
		}

		fmt.Println(vnet.Name)
		return nil
	})
}
