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
	virtualNetworks []*network.VirtualNetwork,
	tags pulumi.StringMap) ([]*compute.BastionHost, error) {

	bastionHostInput := []*BastionHostInput{}
	if err := cfg.TryObject("bastionHosts", &bastionHostInput); err != nil {
		return nil, err
	}

	// this channel is used by the `Output.Apply()` methods to pass values back
	// to the parent goroutine
	applyChan := make(chan bool, 2)
	defer close(applyChan)

	bastionHosts := []*compute.BastionHost{}
	for _, virtualNetwork := range virtualNetworks {
		for _, input := range bastionHostInput {
			virtualNetwork.Name.ApplyBool(func(name string) bool {
				if strings.HasPrefix(name, input.VirtualNetwork) {
					applyChan <- true
					return true
				}

				applyChan <- false
				return false
			})

			if launchBastion := <-applyChan; !launchBastion {
				continue
			}

			subnetID := virtualNetwork.Subnets.ApplyString(func(subnets []network.VirtualNetworkSubnet) string {
				for _, subnet := range subnets {
					if strings.Contains(subnet.Name, "AzureBastionSubnet") {
						return *subnet.Id
					}
				}

				return ""
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
	}

	return bastionHosts, nil
}

type BastionHostInput struct {
	Name           string
	PublicIP       string
	VirtualNetwork string
}
