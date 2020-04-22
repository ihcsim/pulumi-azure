package compute

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
	AppSecGroup           string
	AvailabilitySet       string
	Count                 int
	CustomData            string
	Name                  string
	NetworkInterface      string
	OSProfile             string
	OSProfileLinux        string
	StorageImageReference string
	StorageOSDisk         string
	VirtualNetwork        string
	VMSize                string `json:"vmSize"`
}
