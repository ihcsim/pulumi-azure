package main

import (
	"fmt"

	"github.com/ihcsim/pulumi-azure/v2/config"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/compute"
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

		resourceGroup, err := createResourceGroup(ctx, commonTags)
		if err != nil {
			return err
		}

		virtualNetworks, err := createVirtualNetworks(ctx, resourceGroup, commonTags)
		if err != nil {
			return err
		}

		if _, err := runCompute(ctx, resourceGroup, virtualNetworks, commonTags); err != nil {
			return err
		}

		return nil
	})
}

func createResourceGroup(ctx *pulumi.Context, tags pulumi.StringMap) (*core.ResourceGroup, error) {
	return core.NewResourceGroup(ctx, string(config.ResourceGroup.Name),
		&core.ResourceGroupArgs{
			Location: config.ResourceGroup.Location,
			Tags:     tags,
		})
}

func createVirtualNetworks(
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
				return nil, missingConfigErr{subnet, "subnet"}
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

func runCompute(
	ctx *pulumi.Context,
	resourceGroup *core.ResourceGroup,
	virtualNetworks []*network.VirtualNetwork,
	tags pulumi.StringMap) ([]*compute.VirtualMachine, error) {

	vms := []*compute.VirtualMachine{}
	for _, virtualNetwork := range virtualNetworks {
		for _, vmConfig := range config.VirtualMachines {
			useNetwork := false
			virtualNetwork.Name.ApplyBool(func(name string) bool {
				useNetwork = name == string(vmConfig.VirtualNetwork)
				return useNetwork
			})

			if useNetwork {
				continue
			}

			subnetID := virtualNetwork.Subnets.ApplyString(func(subnets []network.VirtualNetworkSubnet) string {
				for _, subnet := range subnets {
					if subnet.Name == string(vmConfig.Subnet) {
						return *subnet.Id
					}
				}
				return ""
			})

			netInf, err := createPrimaryNetworkInterface(ctx, vmConfig.Name, subnetID, resourceGroup, tags)
			if err != nil {
				return nil, err
			}

			osProfileConfig, exists := config.OSProfiles[vmConfig.OSProfile]
			if !exists {
				return nil, missingConfigErr{vmConfig.OSProfile, "osprofile"}
			}

			osProfile := compute.VirtualMachineOsProfileArgs{
				AdminPassword: &osProfileConfig.AdminPassword,
				AdminUsername: osProfileConfig.AdminUsername,
				ComputerName:  vmConfig.Name,
				CustomData:    osProfileConfig.CustomData,
			}

			osProfileLinuxConfig, exists := config.OSProfileLinux[vmConfig.OSProfileLinux]
			if !exists {
				return nil, missingConfigErr{vmConfig.OSProfileLinux, "osprofile-linux"}
			}

			osProfileLinux := compute.VirtualMachineOsProfileLinuxConfigArgs{
				DisablePasswordAuthentication: osProfileLinuxConfig.DisablePasswordAuthentication,
				SshKeys: compute.VirtualMachineOsProfileLinuxConfigSshKeyArray{
					compute.VirtualMachineOsProfileLinuxConfigSshKeyArgs{
						KeyData: osProfileLinuxConfig.SSHKeyData,
						Path:    osProfileLinuxConfig.SSHKeyPath,
					},
				},
			}

			osDiskConfig, exists := config.StorageOSDisks[vmConfig.StorageOSDisk]
			if !exists {
				return nil, missingConfigErr{vmConfig.StorageOSDisk, "osdisk"}
			}

			osDisk := compute.VirtualMachineStorageOsDiskArgs{
				CreateOption: osDiskConfig.CreateOption,
				DiskSizeGb:   osDiskConfig.DiskSizeGb,
				Name:         vmConfig.Name,
				OsType:       osDiskConfig.OSType,
			}

			imageRefConfig, exists := config.StorageImageReference[vmConfig.StorageImageReference]
			if !exists {
				return nil, missingConfigErr{vmConfig.StorageImageReference, "image-reference"}
			}

			imageRef := compute.VirtualMachineStorageImageReferenceArgs{
				Offer:     imageRefConfig.Offer,
				Publisher: imageRefConfig.Publisher,
				Sku:       imageRefConfig.SKU,
				Version:   imageRefConfig.Version,
			}

			vm, err := compute.NewVirtualMachine(ctx, string(vmConfig.Name), &compute.VirtualMachineArgs{
				Location:                  resourceGroup.Location,
				Name:                      vmConfig.Name,
				OsProfile:                 osProfile,
				OsProfileLinuxConfig:      osProfileLinux,
				PrimaryNetworkInterfaceId: netInf.ID(),
				NetworkInterfaceIds:       pulumi.StringArray{netInf.ID()},
				StorageImageReference:     imageRef,
				ResourceGroupName:         resourceGroup.Name,
				StorageOsDisk:             osDisk,
				Tags:                      tags,
				VmSize:                    vmConfig.VMSize,
			})
			if err != nil {
				return nil, err
			}

			vms = append(vms, vm)
		}
	}

	return vms, nil
}

func createPrimaryNetworkInterface(
	ctx *pulumi.Context,
	virtualMachine pulumi.String,
	subnetID pulumi.StringOutput,
	resourceGroup *core.ResourceGroup,
	tags pulumi.StringMap) (*network.NetworkInterface, error) {

	for netInfKind, netInfConfig := range config.NetworkInterfaces {
		if netInfKind != config.NetworkInterfaceKindPrimary {
			continue
		}

		ipConfigs := network.NetworkInterfaceIpConfigurationArray{}
		for ipConfigKind, ipConfigData := range config.IPConfiguration {
			if ipConfigKind != netInfConfig.IPConfiguration || !ipConfigData.Primary {
				continue
			}

			ipConfigs = append(ipConfigs, network.NetworkInterfaceIpConfigurationArgs{
				Name:                       pulumi.Sprintf("%s-primary-ipconfig", virtualMachine),
				Primary:                    ipConfigData.Primary,
				PrivateIpAddressAllocation: ipConfigData.PrivateIPAddressAllocation,
				PrivateIpAddressVersion:    ipConfigData.PrivateIPAddressVersion,
				SubnetId:                   subnetID,
			})
		}

		args := &network.NetworkInterfaceArgs{
			IpConfigurations:  ipConfigs,
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			Tags:              tags,
		}

		netInfName := fmt.Sprintf("%s-primary", virtualMachine)
		return network.NewNetworkInterface(ctx, netInfName, args)
	}

	return nil, missingConfigErr{virtualMachine, "primary network interface"}
}
