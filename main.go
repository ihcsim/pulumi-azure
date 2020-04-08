package main

import (
	"github.com/ihcsim/pulumi-azure/v2/inputs"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		commonTags := pulumi.StringMap{
			"project": pulumi.String(ctx.Project()),
			"stack":   pulumi.String(ctx.Stack()),
		}

		// create the resource group
		resourceGroup, err := core.NewResourceGroup(ctx, string(inputs.ResourceGroup.Name),
			&core.ResourceGroupArgs{
				Location: inputs.ResourceGroup.Location,
				Tags:     commonTags,
			})
		if err != nil {
			return err
		}

		// create application security groups
		appSecGroups := map[pulumi.String]*network.ApplicationSecurityGroup{}
		for _, secgroup := range inputs.AppSecGroups {
			a, err := network.NewApplicationSecurityGroup(ctx, string(secgroup.Name),
				&network.ApplicationSecurityGroupArgs{
					Location:          resourceGroup.Location,
					Name:              pulumi.String(secgroup.Name),
					ResourceGroupName: resourceGroup.Name,
					Tags:              commonTags,
				})
			if err != nil {
				return err
			}
			appSecGroups[secgroup.Name] = a
		}

		// create network security rules
		networkRules := map[pulumi.String]network.NetworkSecurityGroupSecurityRuleArgs{}
		for _, rule := range inputs.NetworkRules {
			destinationAppSecGroupIds := pulumi.StringArray{}
			for _, secgroup := range rule.DestinationAppSecurityGroups {
				destinationAppSecGroupIds = append(destinationAppSecGroupIds, appSecGroups[secgroup.(pulumi.String)].ID())
			}

			networkRules[rule.Name] = network.NetworkSecurityGroupSecurityRuleArgs{
				Access:                                 rule.Access,
				Description:                            rule.Description,
				DestinationPortRanges:                  rule.DestinationPortRanges,
				DestinationApplicationSecurityGroupIds: destinationAppSecGroupIds,
				Direction:                              rule.Direction,
				Name:                                   rule.Name,
				Priority:                               rule.Priority,
				Protocol:                               rule.Protocol,
				SourceAddressPrefix:                    rule.SourceAddressPrefix,
				SourcePortRange:                        rule.SourcePortRange,
			}
		}

		// create the network security groups
		networkSecGroups := map[pulumi.String]*network.NetworkSecurityGroup{}
		for _, secgroup := range inputs.NetworkSecGroups {
			securityRules := network.NetworkSecurityGroupSecurityRuleArray{}
			for _, rule := range secgroup.SecurityRules {
				securityRules = append(securityRules, networkRules[rule])
			}

			sg, err := network.NewNetworkSecurityGroup(ctx, string(secgroup.Name),
				&network.NetworkSecurityGroupArgs{
					Location:          resourceGroup.Location,
					ResourceGroupName: resourceGroup.Name,
					SecurityRules:     securityRules,
					Tags:              commonTags,
				})
			if err != nil {
				return err
			}

			networkSecGroups[secgroup.Name] = sg
		}

		// create the virtual network
		subnets := map[pulumi.String]network.VirtualNetworkSubnetArgs{}
		for _, subnet := range inputs.Subnets {
			subnets[subnet.Name] = network.VirtualNetworkSubnetArgs{
				AddressPrefix: subnet.AddressPrefix,
				Name:          subnet.Name,
				SecurityGroup: networkSecGroups[subnet.SecurityGroup].ID(),
			}
		}

		for _, vnet := range inputs.VNets {
			subnetsInput := network.VirtualNetworkSubnetArray{}
			for _, subnet := range vnet.Subnets {
				subnetsInput = append(subnetsInput, subnets[subnet])
			}

			if _, err := network.NewVirtualNetwork(ctx, string(vnet.Name),
				&network.VirtualNetworkArgs{
					AddressSpaces:     vnet.CIDR,
					Location:          resourceGroup.Location,
					ResourceGroupName: resourceGroup.Name,
					Tags:              commonTags,
					Subnets:           subnetsInput,
				}); err != nil {
				return err
			}
		}

		return nil
	})
}
