package compute

import (
	"fmt"

	pulumiazure "github.com/ihcsim/pulumi-azure/v2"
	"github.com/ihcsim/pulumi-azure/v2/config"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/compute"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func Up(
	ctx *pulumi.Context,
	resourceGroup *core.ResourceGroup,
	virtualNetworks []*network.VirtualNetwork,
	tags pulumi.StringMap) ([]*compute.VirtualMachine, error) {

	availabilitySets := map[pulumi.String]*compute.AvailabilitySet{}
	for _, asConfig := range config.AvailabilitySets {
		availabilitySet, err := compute.NewAvailabilitySet(ctx, string(asConfig.Name), &compute.AvailabilitySetArgs{
			Location:                  resourceGroup.Location,
			Managed:                   asConfig.Managed,
			Name:                      asConfig.Name,
			PlatformFaultDomainCount:  asConfig.PlatformFaultDomainCount,
			PlatformUpdateDomainCount: asConfig.PlatformUpdateDomainCount,
			ResourceGroupName:         resourceGroup.Name,
			Tags:                      tags,
		})
		if err != nil {
			return nil, err
		}

		availabilitySets[asConfig.Name] = availabilitySet
	}

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
				return nil, pulumiazure.MissingConfigErr{vmConfig.OSProfile, "osprofile"}
			}

			osProfile := compute.VirtualMachineOsProfileArgs{
				AdminPassword: &osProfileConfig.AdminPassword,
				AdminUsername: osProfileConfig.AdminUsername,
				ComputerName:  vmConfig.Name,
				CustomData:    osProfileConfig.CustomData,
			}

			osProfileLinuxConfig, exists := config.OSProfileLinux[vmConfig.OSProfileLinux]
			if !exists {
				return nil, pulumiazure.MissingConfigErr{vmConfig.OSProfileLinux, "osprofile-linux"}
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
				return nil, pulumiazure.MissingConfigErr{vmConfig.StorageOSDisk, "osdisk"}
			}

			osDisk := compute.VirtualMachineStorageOsDiskArgs{
				CreateOption: osDiskConfig.CreateOption,
				DiskSizeGb:   osDiskConfig.DiskSizeGb,
				Name:         vmConfig.Name,
				OsType:       osDiskConfig.OSType,
			}

			imageRefConfig, exists := config.StorageImageReference[vmConfig.StorageImageReference]
			if !exists {
				return nil, pulumiazure.MissingConfigErr{vmConfig.StorageImageReference, "image-reference"}
			}

			imageRef := compute.VirtualMachineStorageImageReferenceArgs{
				Offer:     imageRefConfig.Offer,
				Publisher: imageRefConfig.Publisher,
				Sku:       imageRefConfig.SKU,
				Version:   imageRefConfig.Version,
			}

			availabilitySet, exists := availabilitySets[vmConfig.AvailabilitySet]
			if !exists {
				return nil, pulumiazure.MissingConfigErr{vmConfig.AvailabilitySet, "availability set"}
			}

			vm, err := compute.NewVirtualMachine(ctx, string(vmConfig.Name), &compute.VirtualMachineArgs{
				AvailabilitySetId:         availabilitySet.ID(),
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

	return nil, pulumiazure.MissingConfigErr{virtualMachine, "primary network interface"}
}
