package test

import (
	"fmt"

	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

const (
	ConfigNamespace = "testConfig"
	Location        = "uswest"
	Project         = "testProject"
	Stack           = "testStack"

	AppSecGroupName                           = "test-appsec-group"
	AvailabilitySetName                       = "test-availability-set"
	BastionName                               = "test-bastion"
	IPConfigurationName                       = "test-ip-configuration"
	IPConfigurationPrivateIPAddressAllocation = "Dynamic"
	IPConfigurationPrivateIPAddressVersion    = "IPv4"
	NetworkInterfaceName                      = "test-virtual-machine-00-primary"
	NetworkSecurityRuleName                   = "test-network-rule"
	NetworkSecurityGroupName                  = "test-network-group"
	OSProfileAdminPassword                    = "test-password"
	OSProfileAdminUsername                    = "test-username"
	OSProfileCustomData                       = "test-custom-data"
	OSProfileName                             = "test-osprofile"
	OSProfileLinuxName                        = "test-osprofile-linux"
	OSProfileLinuxSSHKeyData                  = "test-key-data"
	OSProfileLinuxSSHKeyPath                  = "test-key-path"
	PublicIPAllocationMethod                  = "Static"
	PublicIPName                              = "test-public-ip"
	PublicIPSKU                               = "Standard"
	PublicIPVersion                           = "IPv4"
	StorageImageReferenceName                 = "test-storage-image-ref"
	StorageImageReferenceOffer                = "test-storage-image-ref-offer"
	StorageImageReferencePublisher            = "test-storage-image-ref-publisher"
	StorageImageReferenceSKU                  = "test-storage-image-ref-sku"
	StorageImageReferenceVersion              = "test-storage-image-ref-version"
	StorageOSDiskCreateOption                 = "test-storage-os-disk-create-option"
	StorageOSDiskName                         = "test-storage-os-disk"
	StorageOSDiskOSType                       = "test=storage-os-disk-os-type"
	SubnetName                                = "test-subnet"
	ResourceGroupName                         = "test-resource-group"
	VirtualMachineCustomData                  = "test-vm-custom-data"
	VirtualMachineInstanceName                = "test-virtual-machine-00"
	VirtualMachineName                        = "test-virtual-machine"
	VirtualMachineSize                        = "D1_Standard"
	VirtualNetworkName                        = "test-virtual-network"
	VirtualNetworkAddressSpace                = "10.0.0.0/16"
)

var (
	Tags = pulumi.StringMap{
		"key": pulumi.String("value"),
	}
	// Config stores all the mock resources
	Config = map[string]string{
		// mock application security group
		fmt.Sprintf("%s:appSecurityGroups", ConfigNamespace): `
[{
	"name": "` + AppSecGroupName + `"
}]`,

		// mock availability set
		fmt.Sprintf("%s:availabilitySets", ConfigNamespace): `
[{
	"managed": true,
	"name": "` + AvailabilitySetName + `",
	"platformFaultDomainCount": 3,
	"platformUpdateDomainCount": 5
}]`,

		// mock bastion host
		fmt.Sprintf("%s:bastionHosts", ConfigNamespace): `
[{
	"name": "` + BastionName + `",
	"publicIP": "` + PublicIPName + `",
	"subnet": "` + SubnetName + `",
	"virtualNetwork": "` + VirtualNetworkName + `"
}]`,

		// mock IP configuration
		fmt.Sprintf("%s:ipConfiguration", ConfigNamespace): `
[{
	"name": "` + IPConfigurationName + `",
	"primary": true,
	"privateIPAddressAllocation": "` + IPConfigurationPrivateIPAddressAllocation + `",
	"privateIPAddressVersion": "` + IPConfigurationPrivateIPAddressVersion + `"
}]`,
		// mock network interface
		fmt.Sprintf("%s:networkInterfaces", ConfigNamespace): `
[{
	"ipConfiguration": "` + IPConfigurationName + `",
	"name": "` + NetworkInterfaceName + `"
}]`,

		// mock network security rules
		fmt.Sprintf("%s:networkSecurityRules", ConfigNamespace): `
[{
  "access": "Allow",
  "description": "test description",
  "destinationAppSecurityGroups": ["` + AppSecGroupName + `"],
  "destinationPortRanges": ["80"],
  "direction": "Inbound",
  "name": "` + NetworkSecurityRuleName + `",
  "priority": 100,
  "protocol": "Tcp",
  "sourceAddressPrefix": "*",
  "sourcePortRange": "*"
}]`,

		// mock network security groups
		fmt.Sprintf("%s:networkSecurityGroups", ConfigNamespace): `
[{
	"name": "` + NetworkSecurityGroupName + `",
	"securityRules": ["` + NetworkSecurityRuleName + `"]
}]`,

		// mock OS profile
		fmt.Sprintf("%s:osProfiles", ConfigNamespace): `
[{
	"adminPassword": "` + OSProfileAdminPassword + `",
	"adminUsername": "` + OSProfileAdminUsername + `",
	"customData": "` + OSProfileCustomData + `",
	"name": "` + OSProfileName + `"
}]`,

		// mock OS profile Linux
		fmt.Sprintf("%s:osProfilesLinux", ConfigNamespace): `
[{
	"DisablePasswordAuthentication": true,
	"Name": "` + OSProfileLinuxName + `",
	"SSHKeyData": "` + OSProfileLinuxSSHKeyData + `",
	"SSHKeyPath": "` + OSProfileLinuxSSHKeyPath + `"
}]`,

		// mock public IP
		fmt.Sprintf("%s:publicIP", ConfigNamespace): `
[{
	"allocationMethod": "` + PublicIPAllocationMethod + `",
	"ipVersion": "` + PublicIPVersion + `",
	"name": "` + PublicIPName + `",
	"sku": "` + PublicIPSKU + `"
}]`,

		// mock storage image reference
		fmt.Sprintf("%s:storageImageReference", ConfigNamespace): `
[{
	"name": "` + StorageImageReferenceName + `",
	"offer": "` + StorageImageReferenceOffer + `",
	"publisher": "` + StorageImageReferencePublisher + `",
	"sku": "` + StorageImageReferenceSKU + `",
	"version": "` + StorageImageReferenceVersion + `"
}]`,

		// mock storage os disk
		fmt.Sprintf("%s:storageOSDisk", ConfigNamespace): `
[{
	"createOption": "` + StorageOSDiskCreateOption + `",
	"diskSizeGB": 10,
	"name": "` + StorageOSDiskName + `",
	"osType": "` + StorageOSDiskOSType + `"
}]`,

		// mock subnet
		fmt.Sprintf("%s:subnets", ConfigNamespace): `
[{
	"name": "` + SubnetName + `",
	"addressPrefix": "10.0.0.0/24",
	"securityGroup": "` + NetworkSecurityGroupName + `"
}]`,

		fmt.Sprintf("%s:resourceGroup", ConfigNamespace): `
{
	"location": "` + Location + `",
	"name": "` + ResourceGroupName + `"
}`,

		// mock virtual machine
		fmt.Sprintf("%s:virtualMachines", ConfigNamespace): `
[{
	"appSecGroup": "` + AppSecGroupName + `",
	"availabilitySet": "` + AvailabilitySetName + `",
	"count": 3,
	"customData": "` + VirtualMachineCustomData + `",
	"name": "` + VirtualMachineName + `",
	"networkInterface": "` + NetworkInterfaceName + `",
	"osProfile": "` + OSProfileName + `",
	"osProfileLinux": "` + OSProfileLinuxName + `",
	"storageImageReference": "` + StorageImageReferenceName + `",
	"storageOSDisk": "` + StorageOSDiskName + `",
	"virtualNetwork": "` + VirtualNetworkName + `",
	"vmSize": "` + VirtualMachineSize + `"
}]`,

		// mock virtual network
		fmt.Sprintf("%s:virtualNetworks", ConfigNamespace): `
[{
	"name": "` + VirtualNetworkName + `",
	"cidr": "` + VirtualNetworkAddressSpace + `",
	"subnets": ["` + SubnetName + `"]
}]`,
	}
)

func MockApplicationSecurityGroup(ctx *pulumi.Context) (map[string]*network.ApplicationSecurityGroup, error) {
	appSecGroups := map[string]*network.ApplicationSecurityGroup{}
	appSecGroup, err := network.NewApplicationSecurityGroup(ctx, AppSecGroupName, &network.ApplicationSecurityGroupArgs{
		Location:          pulumi.String(Location),
		Name:              pulumi.String(AppSecGroupName),
		ResourceGroupName: pulumi.String(ResourceGroupName),
	})
	if err != nil {
		return nil, err
	}

	appSecGroups[AppSecGroupName] = appSecGroup
	return appSecGroups, nil
}

func MockPublicIPs(ctx *pulumi.Context) (map[string]*network.PublicIp, error) {
	publicIPs := map[string]*network.PublicIp{}
	publicIP, err := network.NewPublicIp(ctx, PublicIPName, &network.PublicIpArgs{
		AllocationMethod:  pulumi.String(PublicIPAllocationMethod),
		IpVersion:         pulumi.String(PublicIPVersion),
		Location:          pulumi.String(Location),
		Name:              pulumi.String(PublicIPName),
		ResourceGroupName: pulumi.String(ResourceGroupName),
		Sku:               pulumi.String(PublicIPSKU),
	})
	if err != nil {
		return nil, err
	}

	publicIPs[PublicIPName] = publicIP
	return publicIPs, nil
}

func MockResourceGroup(ctx *pulumi.Context) (*core.ResourceGroup, error) {
	return core.NewResourceGroup(ctx, ResourceGroupName, &core.ResourceGroupArgs{
		Location: pulumi.String(Location),
		Name:     pulumi.String(ResourceGroupName),
	})
}

func MockVirtualNetworks(ctx *pulumi.Context) (map[string]*network.VirtualNetwork, error) {

	subnet := &network.VirtualNetworkSubnetArgs{
		Name: pulumi.String(SubnetName),
		Id:   pulumi.String(SubnetName + "_id"),
	}

	virtualNetworks := map[string]*network.VirtualNetwork{}
	virtualNetwork, err := network.NewVirtualNetwork(ctx, VirtualNetworkName,
		&network.VirtualNetworkArgs{
			AddressSpaces:     pulumi.StringArray{pulumi.String(VirtualNetworkAddressSpace)},
			Location:          pulumi.String(Location),
			Name:              pulumi.String(VirtualNetworkName),
			ResourceGroupName: pulumi.String(ResourceGroupName),
			Subnets:           network.VirtualNetworkSubnetArray{subnet},
		})
	if err != nil {
		return nil, err
	}

	virtualNetworks[VirtualNetworkName] = virtualNetwork
	return virtualNetworks, nil
}
