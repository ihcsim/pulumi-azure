package network

import (
	pulumiazure "github.com/ihcsim/pulumi-azure/v2"
	"github.com/ihcsim/pulumi-azure/v2/config"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func Up(
	ctx *pulumi.Context,
	resourceGroup *core.ResourceGroup,
	tags pulumi.StringMap) ([]*network.VirtualNetwork, error) {

	appGroups := map[pulumi.String]*network.ApplicationSecurityGroup{}
	for _, appGroupConfig := range config.AppSecGroups {
		appGroup, err := network.NewApplicationSecurityGroup(ctx, string(appGroupConfig.Name),
			&network.ApplicationSecurityGroupArgs{
				Location:          resourceGroup.Location,
				Name:              pulumi.String(appGroupConfig.Name),
				ResourceGroupName: resourceGroup.Name,
				Tags:              tags,
			})
		if err != nil {
			return nil, err
		}

		appGroups[appGroupConfig.Name] = appGroup
	}

	networkRules := map[pulumi.String]network.NetworkSecurityGroupSecurityRuleArgs{}
	for _, ruleConfig := range config.NetworkRules {
		appGroupIDs := pulumi.StringArray{}
		for _, appGroupConfig := range ruleConfig.DestinationAppSecurityGroups {
			appGroupIDs = append(
				appGroupIDs,
				appGroups[appGroupConfig.(pulumi.String)].ID())
		}

		networkRules[ruleConfig.Name] = network.NetworkSecurityGroupSecurityRuleArgs{
			Access:                                 ruleConfig.Access,
			Description:                            ruleConfig.Description,
			DestinationPortRanges:                  ruleConfig.DestinationPortRanges,
			DestinationApplicationSecurityGroupIds: appGroupIDs,
			Direction:                              ruleConfig.Direction,
			Name:                                   ruleConfig.Name,
			Priority:                               ruleConfig.Priority,
			Protocol:                               ruleConfig.Protocol,
			SourceAddressPrefix:                    ruleConfig.SourceAddressPrefix,
			SourcePortRange:                        ruleConfig.SourcePortRange,
		}
	}

	networkSecGroups := map[pulumi.String]*network.NetworkSecurityGroup{}
	for _, secgroupConfig := range config.NetworkSecGroups {
		securityRules := network.NetworkSecurityGroupSecurityRuleArray{}
		for _, ruleConfig := range secgroupConfig.SecurityRules {
			securityRules = append(securityRules, networkRules[ruleConfig])
		}

		securityGroup, err := network.NewNetworkSecurityGroup(ctx, string(secgroupConfig.Name),
			&network.NetworkSecurityGroupArgs{
				Location:          resourceGroup.Location,
				ResourceGroupName: resourceGroup.Name,
				SecurityRules:     securityRules,
				Tags:              tags,
			})
		if err != nil {
			return nil, err
		}

		networkSecGroups[secgroupConfig.Name] = securityGroup
	}

	networks := []*network.VirtualNetwork{}
	for _, vnet := range config.VNets {
		subnets := network.VirtualNetworkSubnetArray{}
		for _, subnet := range vnet.Subnets {
			subnetConfig, exists := config.Subnets[subnet]
			if !exists {
				return nil, pulumiazure.MissingConfigErr{subnet, "subnet"}
			}

			subnets = append(subnets, network.VirtualNetworkSubnetArgs{
				AddressPrefix: subnetConfig.AddressPrefix,
				Name:          subnet,
				SecurityGroup: networkSecGroups[subnetConfig.SecurityGroup].ID(),
			})
		}

		network, err := network.NewVirtualNetwork(ctx, string(vnet.Name),
			&network.VirtualNetworkArgs{
				AddressSpaces:     vnet.CIDR,
				Location:          resourceGroup.Location,
				ResourceGroupName: resourceGroup.Name,
				Tags:              tags,
				Subnets:           subnets,
			})
		if err != nil {
			return nil, err
		}

		networks = append(networks, network)
	}

	return networks, nil
}
