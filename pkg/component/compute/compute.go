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

func Up(
	ctx *pulumi.Context,
	cfg *config.Config,
	resourceGroup *core.ResourceGroup,
	virtualNetworks []*network.VirtualNetwork,
	tags pulumi.StringMap) ([]*compute.VirtualMachine, error) {

	availabilitySets, err := createAvailabilitySets(ctx, cfg, resourceGroup, tags)
	if err != nil {
		return nil, err
	}

	osProfiles, err := createOSProfiles(ctx, cfg)
	if err != nil {
		return nil, err
	}

	osProfilesLinux, err := createOSProfilesLinux(ctx, cfg)
	if err != nil {
		return nil, err
	}

	storageImageReferences, err := createStorageImageReferences(ctx, cfg)
	if err != nil {
		return nil, err
	}

	storageOSDisks, err := createStorageOSDisks(ctx, cfg)
	if err != nil {
		return nil, err
	}

	virtualMachineInput := []*VirtualMachineInput{}
	if err := cfg.TryObject("virtualMachines", &virtualMachineInput); err != nil {
		return nil, err
	}

	virtualMachines := []*compute.VirtualMachine{}
	for _, virtualNetwork := range virtualNetworks {
		for _, vmInput := range virtualMachineInput {
			launchVM := make(chan bool)
			subnetID := virtualNetwork.Subnets.ApplyString(func(subnets []network.VirtualNetworkSubnet) string {
				for _, subnet := range subnets {
					if strings.HasPrefix(subnet.Name, vmInput.Subnet) {
						launchVM <- true
						return *subnet.Id
					}
				}

				launchVM <- false
				return ""
			})

			if t := <-launchVM; !t {
				continue
			}

			osProfile, exists := osProfiles[vmInput.OSProfile]
			if !exists {
				return nil, pulumierr.MissingConfigErr{vmInput.OSProfile, "osprofile"}
			}

			osProfileLinux, exists := osProfilesLinux[vmInput.OSProfileLinux]
			if !exists {
				return nil, pulumierr.MissingConfigErr{vmInput.OSProfileLinux, "osprofile-linux"}
			}

			storageImageReference, exists := storageImageReferences[vmInput.StorageImageReference]
			if !exists {
				return nil, pulumierr.MissingConfigErr{vmInput.StorageImageReference, "storage-image-reference"}
			}

			storageOSDisk, exists := storageOSDisks[vmInput.StorageOSDisk]
			if !exists {
				return nil, pulumierr.MissingConfigErr{vmInput.StorageOSDisk, "storage-os-disk"}
			}
			storageOSDisk.Name = pulumi.String(vmInput.Name)

			availabilitySet, exists := availabilitySets[vmInput.AvailabilitySet]
			if !exists {
				return nil, pulumierr.MissingConfigErr{vmInput.AvailabilitySet, "availability set"}
			}

			var (
				paddingLen         = int(math.Round(float64(vmInput.Count)/10)) + 1
				instanceNamePrefix = fmt.Sprintf("%s-%s", vmInput.Name, strings.Repeat("0", paddingLen))
			)
			for i := 0; i < int(vmInput.Count); i++ {
				instanceName := pulumi.String(fmt.Sprintf("%s%d", instanceNamePrefix, i))
				netInf, err := createPrimaryNetworkInterface(ctx, cfg, instanceName, subnetID, resourceGroup, tags)
				if err != nil {
					return nil, err
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
					VmSize:                    pulumi.String(vmInput.VMSize),
				})
				if err != nil {
					return nil, err
				}

				virtualMachines = append(virtualMachines, virtualMachine)
			}
		}
	}

	return virtualMachines, nil
}

func createAvailabilitySets(
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

func createOSProfiles(
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

func createOSProfilesLinux(
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

func createStorageImageReferences(
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

func createStorageOSDisks(
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

func createPrimaryNetworkInterface(
	ctx *pulumi.Context,
	cfg *config.Config,
	virtualMachine pulumi.String,
	subnetID pulumi.StringOutput,
	resourceGroup *core.ResourceGroup,
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
		return network.NewNetworkInterface(ctx, netInfName, args)
	}

	return nil, pulumierr.MissingConfigErr{string(virtualMachine), "primary network interface"}
}

type AvailabilitySetInput struct {
	Managed                   bool
	Name                      string
	PlatformFaultDomainCount  int
	PlatformUpdateDomainCount int
}

type IPConfigurationInput struct {
	Name                       string
	Primary                    bool
	PrivateIPAddressAllocation string
	PrivateIPAddressVersion    string
}

type NetworkInterfaceInput struct {
	IPConfiguration string `json:"ipConfiguration"`
	Name            string
}

type OSProfileLinuxInput struct {
	DisablePasswordAuthentication bool
	Name                          string
	SSHKeyData                    string
	SSHKeyPath                    string
}

type OSProfileInput struct {
	AdminPassword string
	AdminUsername string
	CustomData    string
	Name          string
}

type StorageImageReferenceInput struct {
	Name      string
	Offer     string
	Publisher string
	SKU       string `json:"sku"`
	Version   string
}

type StorageOSDiskInput struct {
	CreateOption string
	DiskSizeGB   int
	Name         string
	OSType       string
}

type VirtualMachineInput struct {
	AvailabilitySet       string
	Count                 int
	Name                  string
	NetworkInterface      string
	OSProfile             string
	OSProfileLinux        string
	StorageImageReference string
	StorageOSDisk         string
	Subnet                string
	VirtualNetwork        string
	VMSize                string `json:"vmSize"`
}
