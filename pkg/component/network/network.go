package network

import (
	pulumierr "github.com/ihcsim/pulumi-azure/v2/pkg/error"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func Reconcile(
	ctx *pulumi.Context,
	cfg *config.Config,
	appSecGroups map[string]*network.ApplicationSecurityGroup,
	resourceGroup *core.ResourceGroup,
	tags pulumi.StringMap) (map[string]*network.VirtualNetwork, error) {

	networkSecurityRules, err := networkSecurityRules(ctx, cfg, appSecGroups)
	if err != nil {
		return nil, err
	}

	networkSecurityGroups, err := networkSecurityGroups(ctx, cfg, networkSecurityRules, resourceGroup, tags)
	if err != nil {
		return nil, err
	}

	allSubnets, err := subnets(ctx, cfg, networkSecurityGroups)
	if err != nil {
		return nil, err
	}

	virtualNetworkInput := []*VirtualNetworkInput{}
	if err := cfg.TryObject("virtualNetworks", &virtualNetworkInput); err != nil {
		return nil, err
	}

	networks := map[string]*network.VirtualNetwork{}
	for _, input := range virtualNetworkInput {
		subnets := network.VirtualNetworkSubnetArray{}
		for _, input := range input.Subnets {
			subnet, exists := allSubnets[input]
			if !exists {
				return nil, pulumierr.MissingConfigErr{input, "subnet"}
			}
			subnets = append(subnets, subnet)
		}

		addressSpaces := pulumi.StringArray{
			pulumi.String(input.CIDR),
		}

		network, err := network.NewVirtualNetwork(ctx, input.Name,
			&network.VirtualNetworkArgs{
				AddressSpaces:     addressSpaces,
				Location:          resourceGroup.Location,
				ResourceGroupName: resourceGroup.Name,
				Tags:              tags,
				Subnets:           subnets,
			})
		if err != nil {
			return nil, err
		}

		networks[input.Name] = network
	}

	return networks, nil
}

func networkSecurityRules(
	ctx *pulumi.Context,
	cfg *config.Config,
	appSecGroups map[string]*network.ApplicationSecurityGroup) (map[string]network.NetworkSecurityGroupSecurityRuleArgs, error) {

	netSecRulesInput := []*NetworkSecurityRuleInput{}
	if err := cfg.TryObject("networkSecurityRules", &netSecRulesInput); err != nil {
		return nil, err
	}

	networkSecurityRules := map[string]network.NetworkSecurityGroupSecurityRuleArgs{}
	for _, input := range netSecRulesInput {
		destinationAppSecGroups := pulumi.StringArray{}
		for _, key := range input.DestinationAppSecurityGroups {
			appSecGroup, exists := appSecGroups[key]
			if !exists {
				return nil, pulumierr.MissingConfigErr{key, "application security group"}
			}
			destinationAppSecGroups = append(
				destinationAppSecGroups,
				appSecGroup.ID())
		}

		destinationPortRanges := pulumi.StringArray{}
		for _, portRanges := range input.DestinationPortRanges {
			destinationPortRanges = append(
				destinationPortRanges,
				pulumi.String(portRanges))
		}

		networkSecurityRules[input.Name] = network.NetworkSecurityGroupSecurityRuleArgs{
			Access:                                 pulumi.String(input.Access),
			Description:                            pulumi.String(input.Description),
			DestinationPortRanges:                  destinationPortRanges,
			DestinationAddressPrefix:               pulumi.String(input.DestinationAddressPrefix),
			DestinationApplicationSecurityGroupIds: destinationAppSecGroups,
			Direction:                              pulumi.String(input.Direction),
			Name:                                   pulumi.String(input.Name),
			Priority:                               pulumi.Int(input.Priority),
			Protocol:                               pulumi.String(input.Protocol),
			SourceAddressPrefix:                    pulumi.String(input.SourceAddressPrefix),
			SourcePortRange:                        pulumi.String(input.SourcePortRange),
		}
	}

	return networkSecurityRules, nil
}

func networkSecurityGroups(
	ctx *pulumi.Context,
	cfg *config.Config,
	networkSecurityRules map[string]network.NetworkSecurityGroupSecurityRuleArgs,
	resourceGroup *core.ResourceGroup,
	tags pulumi.StringMap) (map[string]pulumi.IDOutput, error) {

	netSecGroupInput := []*NetworkSecurityGroupInput{}
	if err := cfg.TryObject("networkSecurityGroups", &netSecGroupInput); err != nil {
		return nil, err
	}

	networkSecurityGroups := map[string]pulumi.IDOutput{}
	for _, input := range netSecGroupInput {
		securityRules := network.NetworkSecurityGroupSecurityRuleArray{}
		for _, rule := range input.SecurityRules {
			securityRules = append(securityRules, networkSecurityRules[rule])
		}

		securityGroup, err := network.NewNetworkSecurityGroup(ctx, input.Name,
			&network.NetworkSecurityGroupArgs{
				Location:          resourceGroup.Location,
				ResourceGroupName: resourceGroup.Name,
				SecurityRules:     securityRules,
				Tags:              tags,
			})
		if err != nil {
			return nil, err
		}

		networkSecurityGroups[input.Name] = securityGroup.ID()
	}

	return networkSecurityGroups, nil
}

func subnets(
	ctx *pulumi.Context,
	cfg *config.Config,
	networkSecurityGroups map[string]pulumi.IDOutput) (map[string]network.VirtualNetworkSubnetArgs, error) {

	var subnetInput []*SubnetInput
	if err := cfg.TryObject("subnets", &subnetInput); err != nil {
		return nil, err
	}

	subnets := map[string]network.VirtualNetworkSubnetArgs{}
	for _, input := range subnetInput {
		subnets[input.Name] = network.VirtualNetworkSubnetArgs{
			AddressPrefix: pulumi.String(input.AddressPrefix),
			Name:          pulumi.String(input.Name),
			SecurityGroup: networkSecurityGroups[input.SecurityGroup],
		}
	}

	return subnets, nil
}
