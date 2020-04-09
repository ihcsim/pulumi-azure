package config

import "github.com/pulumi/pulumi/sdk/go/pulumi"

var (
	ResourceGroup = struct {
		Location pulumi.String
		Name     pulumi.String
	}{
		Location: "WestUS",
		Name:     "isim-dev",
	}

	VNets = []struct {
		CIDR    pulumi.StringArray
		Name    pulumi.String
		Subnets []pulumi.String
	}{
		{
			CIDR: pulumi.StringArray{pulumi.String("10.0.0.0/16")},
			Name: ResourceGroup.Name,
			Subnets: []pulumi.String{
				Subnets[0].Name,
				Subnets[1].Name,
				Subnets[2].Name,
			},
		},
	}

	Subnets = []struct {
		Name          pulumi.String
		AddressPrefix pulumi.String
		SecurityGroup pulumi.String
	}{
		{
			Name:          "subnet-00",
			AddressPrefix: "10.0.10.0/24",
			SecurityGroup: NetworkSecGroups[0].Name,
		},
		{
			Name:          "subnet-01",
			AddressPrefix: "10.0.20.0/24",
			SecurityGroup: NetworkSecGroups[0].Name,
		},
		{
			Name:          "subnet-02",
			AddressPrefix: "10.0.30.0/24",
			SecurityGroup: NetworkSecGroups[0].Name,
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
				NetworkRules[0].Name,
				NetworkRules[1].Name,
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
				pulumi.String(AppSecGroups[0].Name),
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
				pulumi.String(AppSecGroups[1].Name),
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
