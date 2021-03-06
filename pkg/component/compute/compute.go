package compute

import (
	"fmt"
	"math"
	"strings"

	pulumierr "github.com/ihcsim/pulumi-azure/v2/pkg/error"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/compute"
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
	virtualNetworks map[string]*network.VirtualNetwork,
	tags pulumi.StringMap) (map[string]*compute.VirtualMachine, error) {

	availabilitySets, err := availabilitySets(ctx, cfg, resourceGroup, tags)
	if err != nil {
		return nil, err
	}

	osProfiles, err := osProfiles(ctx, cfg)
	if err != nil {
		return nil, err
	}

	osProfilesLinux, err := osProfilesLinux(ctx, cfg)
	if err != nil {
		return nil, err
	}

	storageImageReferences, err := storageImageReferences(ctx, cfg)
	if err != nil {
		return nil, err
	}

	storageOSDisks, err := storageOSDisks(ctx, cfg)
	if err != nil {
		return nil, err
	}

	virtualMachineInput := []*VirtualMachineInput{}
	if err := cfg.TryObject("virtualMachines", &virtualMachineInput); err != nil {
		return nil, err
	}

	virtualMachines := map[string]*compute.VirtualMachine{}
	for _, input := range virtualMachineInput {
		virtualNetwork, exists := virtualNetworks[input.VirtualNetwork]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.VirtualNetwork, "virtual network"}
		}

		osProfile, exists := osProfiles[input.OSProfile]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.OSProfile, "osprofile"}
		}

		osProfileLinux, exists := osProfilesLinux[input.OSProfileLinux]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.OSProfileLinux, "osprofile-linux"}
		}

		storageImageReference, exists := storageImageReferences[input.StorageImageReference]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.StorageImageReference, "storage-image-reference"}
		}

		storageOSDisk, exists := storageOSDisks[input.StorageOSDisk]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.StorageOSDisk, "storage-os-disk"}
		}
		storageOSDisk.Name = pulumi.String(input.Name)

		availabilitySet, exists := availabilitySets[input.AvailabilitySet]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.AvailabilitySet, "availability set"}
		}

		appSecGroup, exists := appSecGroups[input.AppSecGroup]
		if !exists {
			return nil, pulumierr.MissingConfigErr{input.AppSecGroup, "application security group"}
		}

		var (
			paddingLen         = int(math.Round(float64(input.Count)/10)) + 1
			instanceNamePrefix = fmt.Sprintf("%s-%s", input.Name, strings.Repeat("0", paddingLen))
		)

		for i := 0; i < input.Count; i++ {
			var (
				instanceName = pulumi.String(fmt.Sprintf("%s%d", instanceNamePrefix, i))
				targetSubnet = fmt.Sprintf("subnet-0%d", i)
			)
			subnetID := virtualNetwork.Subnets.ApplyString(func(subnets []network.VirtualNetworkSubnet) (string, error) {
				for _, subnet := range subnets {
					if strings.Contains(subnet.Name, targetSubnet) {
						if subnet.Id == nil {
							return "", pulumierr.MissingConfigErr{targetSubnet, "subnet ID"}
						}
						return *subnet.Id, nil
					}
				}
				return "", nil
			})

			netInf, err := primaryNetworkInterface(ctx, cfg, appSecGroup, resourceGroup, instanceName, subnetID, tags)
			if err != nil {
				return nil, err
			}

			if len(input.CustomData) > 0 {
				osProfile.CustomData = pulumi.Sprintf("%s\n%s", osProfile.CustomData, input.CustomData)
			}

			osProfile.ComputerName = instanceName
			storageOSDisk.Name = instanceName

			virtualMachine, err := compute.NewVirtualMachine(ctx, string(instanceName), &compute.VirtualMachineArgs{
				AvailabilitySetId:         availabilitySet,
				Location:                  resourceGroup.Location,
				Name:                      instanceName,
				OsProfile:                 osProfile,
				OsProfileLinuxConfig:      osProfileLinux,
				PrimaryNetworkInterfaceId: netInf.ID(),
				NetworkInterfaceIds:       pulumi.StringArray{netInf.ID()},
				StorageImageReference:     storageImageReference,
				ResourceGroupName:         resourceGroup.Name,
				StorageOsDisk:             storageOSDisk,
				Tags:                      tags,
				VmSize:                    pulumi.String(input.VMSize),
			})
			if err != nil {
				return nil, err
			}

			virtualMachines[string(instanceName)] = virtualMachine
		}
	}

	return virtualMachines, nil
}

func availabilitySets(
	ctx *pulumi.Context,
	cfg *config.Config,
	resourceGroup *core.ResourceGroup,
	tags pulumi.StringMap) (map[string]pulumi.IDOutput, error) {

	availabilitySetInput := []*AvailabilitySetInput{}
	if err := cfg.TryObject("availabilitySets", &availabilitySetInput); err != nil {
		return nil, err
	}

	availabilitySets := map[string]pulumi.IDOutput{}
	for _, input := range availabilitySetInput {
		availabilitySet, err := compute.NewAvailabilitySet(ctx, input.Name, &compute.AvailabilitySetArgs{
			Location:                  resourceGroup.Location,
			Managed:                   pulumi.Bool(input.Managed),
			Name:                      pulumi.String(input.Name),
			PlatformFaultDomainCount:  pulumi.Int(input.PlatformFaultDomainCount),
			PlatformUpdateDomainCount: pulumi.Int(input.PlatformUpdateDomainCount),
			ResourceGroupName:         resourceGroup.Name,
			Tags:                      tags,
		})
		if err != nil {
			return nil, err
		}

		availabilitySets[input.Name] = availabilitySet.ID()
	}

	return availabilitySets, nil
}

func osProfiles(
	ctx *pulumi.Context,
	cfg *config.Config) (map[string]compute.VirtualMachineOsProfileArgs, error) {

	osProfileInput := []*OSProfileInput{}
	if err := cfg.TryObject("osProfiles", &osProfileInput); err != nil {
		return nil, err
	}

	osProfiles := map[string]compute.VirtualMachineOsProfileArgs{}
	for _, input := range osProfileInput {
		osProfiles[input.Name] = compute.VirtualMachineOsProfileArgs{
			AdminPassword: pulumi.String(input.AdminPassword),
			AdminUsername: pulumi.String(input.AdminUsername),
			CustomData:    pulumi.String(input.CustomData),
		}
	}

	return osProfiles, nil
}

func osProfilesLinux(
	ctx *pulumi.Context,
	cfg *config.Config) (map[string]compute.VirtualMachineOsProfileLinuxConfigArgs, error) {

	osProfileLinuxInput := []*OSProfileLinuxInput{}
	if err := cfg.TryObject("osProfilesLinux", &osProfileLinuxInput); err != nil {
		return nil, err
	}

	osProfilesLinux := map[string]compute.VirtualMachineOsProfileLinuxConfigArgs{}
	for _, input := range osProfileLinuxInput {
		osProfilesLinux[input.Name] = compute.VirtualMachineOsProfileLinuxConfigArgs{
			DisablePasswordAuthentication: pulumi.Bool(input.DisablePasswordAuthentication),
			SshKeys: compute.VirtualMachineOsProfileLinuxConfigSshKeyArray{
				compute.VirtualMachineOsProfileLinuxConfigSshKeyArgs{
					KeyData: pulumi.String(input.SSHKeyData),
					Path:    pulumi.String(input.SSHKeyPath),
				},
			},
		}
	}

	return osProfilesLinux, nil
}

func storageImageReferences(
	ctx *pulumi.Context,
	cfg *config.Config) (map[string]compute.VirtualMachineStorageImageReferenceArgs, error) {

	storageImageReferenceInput := []*StorageImageReferenceInput{}
	if err := cfg.TryObject("storageImageReference", &storageImageReferenceInput); err != nil {
		return nil, err
	}

	storageImageReferences := map[string]compute.VirtualMachineStorageImageReferenceArgs{}
	for _, input := range storageImageReferenceInput {
		storageImageReferences[input.Name] = compute.VirtualMachineStorageImageReferenceArgs{
			Offer:     pulumi.String(input.Offer),
			Publisher: pulumi.String(input.Publisher),
			Sku:       pulumi.String(input.SKU),
			Version:   pulumi.String(input.Version),
		}
	}

	return storageImageReferences, nil
}

func storageOSDisks(
	ctx *pulumi.Context,
	cfg *config.Config) (map[string]compute.VirtualMachineStorageOsDiskArgs, error) {

	storageOSDiskInput := []*StorageOSDiskInput{}
	if err := cfg.TryObject("storageOSDisk", &storageOSDiskInput); err != nil {
		return nil, err
	}

	storageOSDisks := map[string]compute.VirtualMachineStorageOsDiskArgs{}
	for _, input := range storageOSDiskInput {
		storageOSDisks[input.Name] = compute.VirtualMachineStorageOsDiskArgs{
			CreateOption: pulumi.String(input.CreateOption),
			DiskSizeGb:   pulumi.Int(input.DiskSizeGB),
			OsType:       pulumi.String(input.OSType),
		}
	}

	return storageOSDisks, nil
}

func primaryNetworkInterface(
	ctx *pulumi.Context,
	cfg *config.Config,
	appSecGroup *network.ApplicationSecurityGroup,
	resourceGroup *core.ResourceGroup,
	virtualMachine pulumi.String,
	subnetID pulumi.StringOutput,
	tags pulumi.StringMap) (*network.NetworkInterface, error) {

	networkInterfaceInput := []*NetworkInterfaceInput{}
	if err := cfg.TryObject("networkInterfaces", &networkInterfaceInput); err != nil {
		return nil, err
	}

	ipConfigurationInput := []*IPConfigurationInput{}
	if err := cfg.TryObject("ipConfiguration", &ipConfigurationInput); err != nil {
		return nil, err
	}

	for _, infInput := range networkInterfaceInput {
		var ipConfigs network.NetworkInterfaceIpConfigurationArray

		for _, ipConfigInput := range ipConfigurationInput {
			if ipConfigInput.Name != infInput.IPConfiguration ||
				(ipConfigInput.Name == infInput.IPConfiguration && !ipConfigInput.Primary) {
				continue
			}

			ipConfigs = append(ipConfigs, network.NetworkInterfaceIpConfigurationArgs{
				Name:                       pulumi.Sprintf("%s-primary-ipconfig", virtualMachine),
				Primary:                    pulumi.Bool(ipConfigInput.Primary),
				PrivateIpAddressAllocation: pulumi.String(ipConfigInput.PrivateIPAddressAllocation),
				PrivateIpAddressVersion:    pulumi.String(ipConfigInput.PrivateIPAddressVersion),
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
		netInf, err := network.NewNetworkInterface(ctx, netInfName, args)
		if err != nil {
			return nil, err
		}

		if _, err := network.NewNetworkInterfaceApplicationSecurityGroupAssociation(ctx, netInfName,
			&network.NetworkInterfaceApplicationSecurityGroupAssociationArgs{
				ApplicationSecurityGroupId: appSecGroup.ID(),
				NetworkInterfaceId:         netInf.ID(),
			}); err != nil {
			return nil, err
		}

		return netInf, nil
	}

	return nil, pulumierr.MissingConfigErr{string(virtualMachine), "primary network interface"}
}
