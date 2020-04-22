package loadbalancer

import (
	"fmt"
	"strings"

	azurenetwork "github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	pulumierr "github.com/ihcsim/pulumi-azure/v2/pkg/error"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/lb"
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
	tags pulumi.StringMap) (map[string]*lb.LoadBalancer, error) {

	loadBalancerInput := []*LoadBalancerInput{}
	if err := cfg.TryObject("loadBalancers", &loadBalancerInput); err != nil {
		return nil, err
	}

	loadBalancers := map[string]*lb.LoadBalancer{}
	for _, input := range loadBalancerInput {
		publicIP, exists := publicIPs[input.PublicIP]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.PublicIP, "public IP"}
		}

		frontendIPConfigurationName := fmt.Sprintf("%s-frontend-config", input.Name)
		frontendIPConfiguration := &lb.LoadBalancerFrontendIpConfigurationArgs{
			Name:              pulumi.String(frontendIPConfigurationName),
			PublicIpAddressId: publicIP.ID(),
		}
		frontendIPConfigurations := lb.LoadBalancerFrontendIpConfigurationArray{frontendIPConfiguration}

		loadBalancer, err := lb.NewLoadBalancer(ctx, input.Name,
			&lb.LoadBalancerArgs{
				FrontendIpConfigurations: frontendIPConfigurations,
				Location:                 resourceGroup.Location,
				Name:                     pulumi.String(input.Name),
				ResourceGroupName:        resourceGroup.Name,
				Sku:                      pulumi.String(input.SKU),
				Tags:                     tags,
			})
		if err != nil {
			return nil, err
		}

		loadBalancers[input.Name] = loadBalancer

		backendAddressPoolName := fmt.Sprintf("%s-backend-pool", input.Name)
		backendAddressPool, err := lb.NewBackendAddressPool(ctx, backendAddressPoolName, &lb.BackendAddressPoolArgs{
			LoadbalancerId:    loadBalancer.ID(),
			Name:              pulumi.String(backendAddressPoolName),
			ResourceGroupName: resourceGroup.Name,
		})
		if err != nil {
			return nil, err
		}

		virtualNetwork, exists := virtualNetworks[input.VirtualNetwork]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.VirtualNetwork, "virtual network"}
		}
		subnetID := virtualNetwork.Subnets.ApplyString(func(subnets []network.VirtualNetworkSubnet) (string, error) {
			for _, subnet := range subnets {
				if strings.Contains(subnet.Name, input.Subnet) {
					if subnet.Id == nil {
						return "", pulumierr.MissingConfigErr{input.Subnet, "subnet ID"}
					}

					return *subnet.Id, nil
				}
			}

			return "", pulumierr.MissingConfigErr{input.Subnet, "subnet"}
		})

		networkInterfaceName := fmt.Sprintf("%s-netinf", input.Name)
		networkInterface, err := network.NewNetworkInterface(ctx, networkInterfaceName,
			&network.NetworkInterfaceArgs{
				IpConfigurations: network.NetworkInterfaceIpConfigurationArray{
					network.NetworkInterfaceIpConfigurationArgs{
						Name:                       pulumi.String(networkInterfaceName),
						Primary:                    pulumi.Bool(true),
						PrivateIpAddressAllocation: pulumi.String(azurenetwork.Dynamic),
						PrivateIpAddressVersion:    pulumi.String(azurenetwork.IPv4),
						SubnetId:                   subnetID,
					},
				},
				Location:          resourceGroup.Location,
				Name:              pulumi.String(networkInterfaceName),
				ResourceGroupName: resourceGroup.Name,
				Tags:              tags,
			})
		if err != nil {
			return nil, err
		}

		associationName := fmt.Sprintf("%s-backend-network-interface", input.Name)
		if _, err := network.NewNetworkInterfaceBackendAddressPoolAssociation(ctx, associationName,
			&network.NetworkInterfaceBackendAddressPoolAssociationArgs{
				BackendAddressPoolId: backendAddressPool.ID(),
				IpConfigurationName:  networkInterface.IpConfigurations.Index(pulumi.Int(0)).Name(),
				NetworkInterfaceId:   networkInterface.ID(),
			}); err != nil {
			return nil, err
		}

		probeName := fmt.Sprintf("%s-probe-web", input.Name)
		probe, err := lb.NewProbe(ctx, probeName, &lb.ProbeArgs{
			LoadbalancerId:    loadBalancer.ID(),
			Name:              pulumi.String(probeName),
			Port:              pulumi.Int(input.ProbePort),
			Protocol:          pulumi.String(input.ProbeProtocol),
			RequestPath:       pulumi.String(input.ProbeRequestPath),
			ResourceGroupName: resourceGroup.Name,
		})
		if err != nil {
			return nil, err
		}

		webRuleName := fmt.Sprintf("%s-rule-web", input.Name)
		if _, err := lb.NewRule(ctx, webRuleName, &lb.RuleArgs{
			BackendAddressPoolId:        backendAddressPool.ID(),
			BackendPort:                 pulumi.Int(input.BackendPort),
			FrontendIpConfigurationName: frontendIPConfiguration.Name,
			FrontendPort:                pulumi.Int(input.FrontendPort),
			LoadbalancerId:              loadBalancer.ID(),
			Name:                        pulumi.String(webRuleName),
			ProbeId:                     probe.ID(),
			Protocol:                    pulumi.String(input.Protocol),
			ResourceGroupName:           resourceGroup.Name,
		}); err != nil {
			return nil, err
		}
	}

	return loadBalancers, nil
}
