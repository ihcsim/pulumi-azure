package network

type NetworkSecurityGroupInput struct {
	Name          string
	SecurityRules []string
}

type NetworkSecurityRuleInput struct {
	Access                       string
	Description                  string
	DestinationAddressPrefix     string
	DestinationAppSecurityGroups []string
	DestinationPortRanges        []string
	Direction                    string
	Name                         string
	Priority                     int
	Protocol                     string
	SourceAddressPrefix          string
	SourcePortRange              string
}

type SubnetInput struct {
	AddressPrefix string
	Name          string
	SecurityGroup string
}

type VirtualNetworkInput struct {
	CIDR    string
	Name    string
	Subnets []string
}
