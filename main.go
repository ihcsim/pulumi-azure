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
		name   string
		owners []string
	}{
		{
			name:   "default",
			owners: []string{string(subnetMeta[0].name), string(subnetMeta[1].name), string(subnetMeta[2].name)},
		},
	}

	networkRules = []struct {
		access                       pulumi.String
		description                  pulumi.String
		destinationPortRanges        pulumi.StringArray
		destinationAppSecurityGroups pulumi.StringArray
		direction                    pulumi.String
		name                         pulumi.String
		owners                       pulumi.StringArray // network security groups that own this rule
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
			owners:                       pulumi.StringArray{pulumi.String(networkSecGroupMeta[0].name)},
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
			owners:                       pulumi.StringArray{pulumi.String(networkSecGroupMeta[0].name)},
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

		// create network security rules.
		// every rule will be stored as value in the `networkSecRules` map keyed
		// off the rule owner's name. The rule owner is a network security group.
		networkSecRules := map[pulumi.String]network.NetworkSecurityGroupSecurityRuleArray{}
		for _, rule := range networkRules {
			// bind the application security groups to the network rules using the
			// `appSecGroups` map created above
			appSecGroupIDs := pulumi.StringArray{}
			for _, appID := range rule.destinationAppSecurityGroups {
				appSecGroupIDs = append(appSecGroupIDs, appSecGroups[appID.(pulumi.String)].ID())
			}

			for _, secgroup := range rule.owners {
				rules, exists := networkSecRules[secgroup.(pulumi.String)]
				if !exists {
					networkSecRules[secgroup.(pulumi.String)] = network.NetworkSecurityGroupSecurityRuleArray{}
				}

				networkSecRules[secgroup.(pulumi.String)] = append(rules,
					network.NetworkSecurityGroupSecurityRuleArgs{
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
					})
			}
		}

		// create the network security groups
		networkSecGroups := map[pulumi.String]*network.NetworkSecurityGroup{}
		for _, meta := range networkSecGroupMeta {
			secGroup, err := network.NewNetworkSecurityGroup(ctx, meta.name,
				&network.NetworkSecurityGroupArgs{
					Location:          resourceGroup.Location,
					ResourceGroupName: resourceGroup.Name,
					SecurityRules:     networkSecRules[pulumi.String(meta.name)],
					Tags:              commonTags,
				})
			if err != nil {
				return err
			}

			for _, owner := range meta.owners {
				networkSecGroups[pulumi.String(owner)] = secGroup
			}
		}

		// create the virtual subnets
		subnets := network.VirtualNetworkSubnetArray{}
		for _, meta := range subnetMeta {
			subnets = append(subnets, network.VirtualNetworkSubnetArgs{
				AddressPrefix: meta.addressPrefix,
				Name:          meta.name,
				SecurityGroup: networkSecGroups[meta.name].ID(),
			})
		}

		// create the virtual network
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
