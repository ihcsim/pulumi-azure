package inputs

import "github.com/pulumi/pulumi/sdk/go/pulumi"

var (
	VNets = struct {
		CIDR pulumi.StringArray
	}{
		CIDR: pulumi.StringArray{pulumi.String("10.0.0.0/16")},
	}

	Subnets = []struct {
		Name          pulumi.String
		AddressPrefix pulumi.String
		SecurityGroup pulumi.String
	}{
		{Name: pulumi.String("subnet-00"), AddressPrefix: pulumi.String("10.0.10.0/24")},
		{Name: pulumi.String("subnet-01"), AddressPrefix: pulumi.String("10.0.20.0/24")},
		{Name: pulumi.String("subnet-02"), AddressPrefix: pulumi.String("10.0.30.0/24")},
	}

	AppSecGroups = []struct {
		Name string
	}{
		{Name: "web-servers"},
		{Name: "admin-servers"},
	}

	NetworkSecGroups = []struct {
		Name   string
		Owners []string
	}{
		{
			Name: "default",
			Owners: []string{
				string(Subnets[0].Name),
				string(Subnets[1].Name),
				string(Subnets[2].Name),
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
		Owners                       pulumi.StringArray // network security groups that own this rule
		Priority                     pulumi.Int
		Protocol                     pulumi.String
		SourceAddressPrefix          pulumi.String
		SourcePortRange              pulumi.String
	}{
		{
			Access:                       pulumi.String("Allow"),
			Description:                  pulumi.String("allow HTTP and HTTPS"),
			DestinationPortRanges:        pulumi.StringArray{pulumi.String("80"), pulumi.String("443")},
			DestinationAppSecurityGroups: pulumi.StringArray{pulumi.String(AppSecGroups[0].Name)},
			Direction:                    "Inbound",
			Name:                         pulumi.String("allow-web-all"),
			Owners:                       pulumi.StringArray{pulumi.String(NetworkSecGroups[0].Name)},
			Priority:                     100,
			Protocol:                     pulumi.String("Tcp"),
			SourceAddressPrefix:          "AzureLoadBalancer",
			SourcePortRange:              "*",
		},
		{
			Access:                       pulumi.String("Allow"),
			Description:                  pulumi.String("allow SSH"),
			DestinationPortRanges:        pulumi.StringArray{pulumi.String("22")},
			DestinationAppSecurityGroups: pulumi.StringArray{pulumi.String(AppSecGroups[1].Name)},
			Direction:                    "Inbound",
			Name:                         pulumi.String("allow-ssh-all"),
			Owners:                       pulumi.StringArray{pulumi.String(NetworkSecGroups[0].Name)},
			Priority:                     101,
			Protocol:                     pulumi.String("Tcp"),
			SourceAddressPrefix:          "*",
			SourcePortRange:              "*",
		},
	}
)
