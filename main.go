package main

import (
	"fmt"

	"github.com/ihcsim/pulumi-azure/v2/inputs"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

const (
	owner  = "isim-dev"
	region = pulumi.String("WestUS")
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		commonTags := pulumi.StringMap{
			"project": pulumi.String(ctx.Project()),
			"stack":   pulumi.String(ctx.Stack()),
		}

		// create the resource group
		resourceGroup, err := core.NewResourceGroup(ctx, owner,
			&core.ResourceGroupArgs{
				Location: region,
				Tags:     commonTags,
			})
		if err != nil {
			return err
		}

		// create application security groups.
		// each security group is stored as a value in the `appSecGroups` map keyed
		// off the security group's name.
		// the `appSecGroups` map will be used later to bind the application
		// security groups to the network security rules.
		appSecGroups := map[pulumi.String]*network.ApplicationSecurityGroup{}
		for _, secgroup := range inputs.AppSecGroups {
			sg, err := network.NewApplicationSecurityGroup(ctx, secgroup.Name,
				&network.ApplicationSecurityGroupArgs{
					Location:          resourceGroup.Location,
					Name:              pulumi.String(secgroup.Name),
					ResourceGroupName: resourceGroup.Name,
					Tags:              commonTags,
				})
			if err != nil {
				return err
			}

			appSecGroups[pulumi.String(secgroup.Name)] = sg
		}

		// create network security rules.
		// every rule will be stored as value in the `networkSecRules` map keyed
		// off the rule owner's name. The rule owner is a network security group.
		networkSecRules := map[pulumi.String]network.NetworkSecurityGroupSecurityRuleArray{}
		for _, rule := range inputs.NetworkRules {
			// bind the application security groups to the network rules using the
			// `appSecGroups` map created above
			appSecGroupIDs := pulumi.StringArray{}
			for _, appID := range rule.DestinationAppSecurityGroups {
				appSecGroupIDs = append(appSecGroupIDs, appSecGroups[appID.(pulumi.String)].ID())
			}

			for _, secgroup := range rule.Owners {
				rules, exists := networkSecRules[secgroup.(pulumi.String)]
				if !exists {
					networkSecRules[secgroup.(pulumi.String)] = network.NetworkSecurityGroupSecurityRuleArray{}
				}

				networkSecRules[secgroup.(pulumi.String)] = append(rules,
					network.NetworkSecurityGroupSecurityRuleArgs{
						Access:                                 rule.Access,
						Description:                            rule.Description,
						DestinationPortRanges:                  rule.DestinationPortRanges,
						DestinationApplicationSecurityGroupIds: appSecGroupIDs,
						Direction:                              rule.Direction,
						Name:                                   rule.Name,
						Priority:                               rule.Priority,
						Protocol:                               rule.Protocol,
						SourceAddressPrefix:                    rule.SourceAddressPrefix,
						SourcePortRange:                        rule.SourcePortRange,
					})
			}
		}

		// create the network security groups
		networkSecGroups := map[pulumi.String]*network.NetworkSecurityGroup{}
		for _, secgroup := range inputs.NetworkSecGroups {
			sg, err := network.NewNetworkSecurityGroup(ctx, secgroup.Name,
				&network.NetworkSecurityGroupArgs{
					Location:          resourceGroup.Location,
					ResourceGroupName: resourceGroup.Name,
					SecurityRules:     networkSecRules[pulumi.String(secgroup.Name)],
					Tags:              commonTags,
				})
			if err != nil {
				return err
			}

			for _, owner := range secgroup.Owners {
				networkSecGroups[pulumi.String(owner)] = sg
			}
		}

		// create the virtual subnets
		subnets := network.VirtualNetworkSubnetArray{}
		for _, meta := range inputs.Subnets {
			subnets = append(subnets, network.VirtualNetworkSubnetArgs{
				AddressPrefix: meta.AddressPrefix,
				Name:          meta.Name,
				SecurityGroup: networkSecGroups[meta.Name].ID(),
			})
		}

		// create the virtual network
		vnet, err := network.NewVirtualNetwork(ctx, owner,
			&network.VirtualNetworkArgs{
				AddressSpaces:     inputs.VNets.CIDR,
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
