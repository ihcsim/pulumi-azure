package config

import "github.com/pulumi/pulumi/sdk/go/pulumi"

const projectName = "isim-dev"

var (
	ResourceGroup = struct {
		Location pulumi.String
		Name     pulumi.String
	}{
		Location: "WestUS",
		Name:     projectName,
	}

	VNets = map[pulumi.String]struct {
		CIDR    pulumi.StringArray
		Name    pulumi.String
		Subnets []pulumi.String
	}{
		projectName: {
			CIDR: pulumi.StringArray{
				pulumi.String("10.0.0.0/16"),
			},
			Name: ResourceGroup.Name,
			Subnets: []pulumi.String{
				"subnet-00",
				"subnet-01",
				"subnet-02",
			},
		},
	}

	Subnets = map[pulumi.String]struct {
		Name          pulumi.String
		AddressPrefix pulumi.String
		SecurityGroup pulumi.String
	}{
		"subnet-00": {
			Name:          "subnet-00",
			AddressPrefix: "10.0.10.0/24",
			SecurityGroup: "default",
		},
		"subnet-01": {
			Name:          "subnet-01",
			AddressPrefix: "10.0.20.0/24",
			SecurityGroup: "default",
		},
		"subnet-02": {
			Name:          "subnet-02",
			AddressPrefix: "10.0.30.0/24",
			SecurityGroup: "default",
		},
	}

	AppSecGroups = []struct {
		Name pulumi.String
	}{
		{Name: "web-servers"},
		{Name: "admin-servers"},
	}

	NetworkSecGroups = []struct {
		Name          pulumi.String
		SecurityRules []pulumi.String
	}{
		{
			Name: "default",
			SecurityRules: []pulumi.String{
				"allow-web-all",
				"allow-ssh-all",
			},
		},
	}

	NetworkRules = []struct {
		Access                       pulumi.String
		Description                  pulumi.String
		DestinationPortRanges        pulumi.StringArray
		DestinationAppSecurityGroups pulumi.StringArray
		Direction                    pulumi.String
		Name                         pulumi.String
		Priority                     pulumi.Int
		Protocol                     pulumi.String
		SourceAddressPrefix          pulumi.String
		SourcePortRange              pulumi.String
	}{
		{
			Access:      "Allow",
			Description: "allow HTTP and HTTPS",
			DestinationPortRanges: pulumi.StringArray{
				pulumi.String("80"),
				pulumi.String("443"),
			},
			DestinationAppSecurityGroups: pulumi.StringArray{
				pulumi.String("web-servers"),
			},
			Direction:           "Inbound",
			Name:                "allow-web-all",
			Priority:            100,
			Protocol:            "Tcp",
			SourceAddressPrefix: "AzureLoadBalancer",
			SourcePortRange:     "*",
		},
		{
			Access:      "Allow",
			Description: "allow SSH",
			DestinationPortRanges: pulumi.StringArray{
				pulumi.String("22"),
			},
			DestinationAppSecurityGroups: pulumi.StringArray{
				pulumi.String("admin-servers"),
			},
			Direction:           "Inbound",
			Name:                "allow-ssh-all",
			Priority:            101,
			Protocol:            "Tcp",
			SourceAddressPrefix: "*",
			SourcePortRange:     "*",
		},
	}
)
