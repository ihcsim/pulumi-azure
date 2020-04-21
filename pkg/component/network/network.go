package network

import (
	pulumierr "github.com/ihcsim/pulumi-azure/v2/pkg/error"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func Up(
	ctx *pulumi.Context,
	cfg *config.Config,
	resourceGroup *core.ResourceGroup,
	tags pulumi.StringMap) ([]*network.VirtualNetwork, error) {

	appSecGroupIDs, err := createApplicationSecurityGroups(ctx, cfg, resourceGroup, tags)
	if err != nil {
		return nil, err
	}

	networkSecurityRules, err := createNetworkSecurityRules(ctx, cfg, appSecGroupIDs)
	if err != nil {
		return nil, err
	}

	networkSecurityGroups, err := createNetworkSecurityGroups(ctx, cfg, networkSecurityRules, resourceGroup, tags)
	if err != nil {
		return nil, err
	}

	allSubnets, err := createSubnets(ctx, cfg, networkSecurityGroups)
	if err != nil {
		return nil, err
	}

	virtualNetworkInput := []*VirtualNetworkInput{}
	if err := cfg.TryObject("virtualNetworks", &virtualNetworkInput); err != nil {
		return nil, err
	}

	networks := []*network.VirtualNetwork{}
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

		networks = append(networks, network)
	}

	return networks, nil
}

func createApplicationSecurityGroups(
	ctx *pulumi.Context,
	cfg *config.Config,
	resourceGroup *core.ResourceGroup,
	tags pulumi.StringMap) (map[string]pulumi.IDOutput, error) {

	appSecGroupsInput := []*ApplicationSecurityGroupInput{}
	if err := cfg.TryObject("appSecurityGroups", &appSecGroupsInput); err != nil {
		return nil, err
	}

	appSecGroupIDs := map[string]pulumi.IDOutput{}
	for _, input := range appSecGroupsInput {
		appSecGroup, err := network.NewApplicationSecurityGroup(ctx, input.Name,
			&network.ApplicationSecurityGroupArgs{
				Location:          resourceGroup.Location,
				Name:              pulumi.String(input.Name),
				ResourceGroupName: resourceGroup.Name,
				Tags:              tags,
			})
		if err != nil {
			return nil, err
		}

		appSecGroupIDs[input.Name] = appSecGroup.ID()
	}

	return appSecGroupIDs, nil
}

func createNetworkSecurityRules(
	ctx *pulumi.Context,
	cfg *config.Config,
	appSecGroupIDs map[string]pulumi.IDOutput) (map[string]network.NetworkSecurityGroupSecurityRuleArgs, error) {

	netSecRulesInput := []*NetworkSecurityRuleInput{}
	if err := cfg.TryObject("networkSecurityRules", &netSecRulesInput); err != nil {
		return nil, err
	}

	networkSecurityRules := map[string]network.NetworkSecurityGroupSecurityRuleArgs{}
	for _, input := range netSecRulesInput {
		destinationAppSecGroups := pulumi.StringArray{}
		for _, key := range input.DestinationAppSecurityGroups {
			id, exists := appSecGroupIDs[key]
			if !exists {
				return nil, pulumierr.MissingConfigErr{key, "application security group"}
			}
			destinationAppSecGroups = append(
				destinationAppSecGroups,
				id)
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

func createNetworkSecurityGroups(
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

func createSubnets(
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

type ApplicationSecurityGroupInput struct {
	Name string
}

type NetworkSecurityGroupInput struct {
	Name          string
	SecurityRules []string
}

type NetworkSecurityRuleInput struct {
	Access                       string
	Description                  string
	DestinationAddressPrefix     string
	DestinationAppSecurityGroups []string
	DestinationPortRanges        []string
	Direction                    string
	Name                         string
	Priority                     int
	Protocol                     string
	SourceAddressPrefix          string
	SourcePortRange              string
}

type SubnetInput struct {
	AddressPrefix string
	Name          string
	SecurityGroup string
}

type VirtualNetworkInput struct {
	CIDR    string
	Name    string
	Subnets []string
}
