package bastion

import (
	"strings"

	pulumierr "github.com/ihcsim/pulumi-azure/v2/pkg/error"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/compute"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func Up(
	ctx *pulumi.Context,
	cfg *config.Config,
	publicIPs map[string]*network.PublicIp,
	resourceGroup *core.ResourceGroup,
	virtualNetworks map[string]*network.VirtualNetwork,
	tags pulumi.StringMap) ([]*compute.BastionHost, error) {

	bastionHostInput := []*BastionHostInput{}
	if err := cfg.TryObject("bastionHosts", &bastionHostInput); err != nil {
		return nil, err
	}

	bastionHosts := []*compute.BastionHost{}
	for _, input := range bastionHostInput {
		virtualNetwork, exists := virtualNetworks[input.VirtualNetwork]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.VirtualNetwork, "virtual network"}
		}

		targetSubnet := "AzureBastionSubnet"
		subnetID := virtualNetwork.Subnets.ApplyString(func(subnets []network.VirtualNetworkSubnet) (string, error) {
			for _, subnet := range subnets {
				if strings.Contains(subnet.Name, targetSubnet) {
					if subnet.Id == nil {
						return "", pulumierr.MissingConfigErr{targetSubnet, "subnet ID"}
					}

					return *subnet.Id, nil
				}
			}

			return "", pulumierr.MissingConfigErr{targetSubnet, "subnet"}
		})

		publicIP, exists := publicIPs[input.PublicIP]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.PublicIP, "public-ip"}
		}

		bastionHost, err := compute.NewBastionHost(ctx, input.Name,
			&compute.BastionHostArgs{
				IpConfiguration: compute.BastionHostIpConfigurationArgs{
					Name:              pulumi.String(input.Name),
					PublicIpAddressId: publicIP.ID(),
					SubnetId:          subnetID,
				},
				Location:          resourceGroup.Location,
				Name:              pulumi.String(input.Name),
				ResourceGroupName: resourceGroup.Name,
				Tags:              tags,
			})
		if err != nil {
			return nil, err
		}

		bastionHosts = append(bastionHosts, bastionHost)
	}

	return bastionHosts, nil
}
