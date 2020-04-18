package config

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

const NetworkInterfaceKindPrimary = "primary"

var (
	AvailabilitySets = []struct {
		Managed                   pulumi.Bool
		Name                      pulumi.String
		PlatformFaultDomainCount  pulumi.Int
		PlatformUpdateDomainCount pulumi.Int
	}{
		{
			Managed:                   true,
			Name:                      "web",
			PlatformFaultDomainCount:  3,
			PlatformUpdateDomainCount: 3,
		},
		{
			Managed:                   true,
			Name:                      "backend",
			PlatformFaultDomainCount:  3,
			PlatformUpdateDomainCount: 3,
		},
	}

	VirtualMachines = []struct {
		AvailabilitySet       pulumi.String
		Count                 pulumi.Int
		Name                  pulumi.String
		NetworkInterface      pulumi.String
		OSProfile             pulumi.String
		OSProfileLinux        pulumi.String
		StorageImageReference pulumi.String
		StorageOSDisk         pulumi.String
		VirtualNetwork        pulumi.String
		Subnet                pulumi.String
		VMSize                pulumi.String
	}{
		{
			AvailabilitySet:       "web",
			Count:                 3,
			Name:                  "web",
			NetworkInterface:      NetworkInterfaceKindPrimary,
			OSProfile:             "default",
			OSProfileLinux:        "default",
			StorageImageReference: "ubuntu-16.04",
			StorageOSDisk:         "default",
			VMSize:                "Standard_B1ls",
			VirtualNetwork:        projectName,
			Subnet:                "subnet-00",
		},
		{
			AvailabilitySet:       "backend",
			Count:                 3,
			Name:                  "backend",
			NetworkInterface:      NetworkInterfaceKindPrimary,
			OSProfile:             "default",
			OSProfileLinux:        "default",
			StorageImageReference: "ubuntu-16.04",
			StorageOSDisk:         "default",
			VMSize:                "Standard_B1ls",
			VirtualNetwork:        projectName,
			Subnet:                "subnet-00",
		},
	}

	IPConfiguration = map[pulumi.String]struct {
		Primary                    pulumi.Bool
		PrivateIPAddressAllocation pulumi.String
		PrivateIPAddressVersion    pulumi.String
	}{
		"ipv4-private-dynamic": {
			Primary:                    true,
			PrivateIPAddressAllocation: pulumi.String(network.Dynamic),
			PrivateIPAddressVersion:    pulumi.String(network.IPv4),
		},
	}

	NetworkInterfaces = map[pulumi.String]struct {
		IPConfiguration pulumi.String
	}{
		NetworkInterfaceKindPrimary: {
			IPConfiguration: "ipv4-private-dynamic",
		},
	}

	OSProfiles = map[pulumi.String]struct {
		AdminPassword pulumi.String
		AdminUsername pulumi.String
		CustomData    pulumi.String
	}{
		"default": {
			AdminPassword: "",
			AdminUsername: "",
			CustomData:    "apt install -y ntpd",
		},
	}

	OSProfileLinux = map[pulumi.String]struct {
		DisablePasswordAuthentication pulumi.Bool
		Name                          pulumi.String
		SSHKeyData                    pulumi.String
		SSHKeyPath                    pulumi.String
	}{
		"default": {
			DisablePasswordAuthentication: false,
			Name:                          "default",
			SSHKeyData:                    "",
			SSHKeyPath:                    "",
		},
	}

	StorageImageReference = map[pulumi.String]struct {
		Offer     pulumi.String
		Publisher pulumi.String
		SKU       pulumi.String
		Version   pulumi.String
	}{
		"ubuntu-16.04": {
			Offer:     "UbuntuServer",
			Publisher: "Canonical",
			SKU:       "16.04-LTS",
			Version:   "latest",
		},
	}

	StorageOSDisks = map[pulumi.String]struct {
		CreateOption pulumi.String
		DiskSizeGb   pulumi.Int
		OSType       pulumi.String
	}{
		"default": {
			CreateOption: pulumi.String(compute.DiskCreateOptionTypesFromImage),
			DiskSizeGb:   30,
			OSType:       pulumi.String(compute.Linux),
		},
	}
)
