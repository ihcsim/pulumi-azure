package loadbalancer

type LoadBalancerInput struct {
	BackendPort      int
	BackendHosts     []string
	FrontendPort     int
	Name             string
	ProbePort        int
	ProbeProtocol    string
	ProbeRequestPath string
	Protocol         string
	PublicIP         string
	SKU              string `json:"sku"`
	Subnet           string
	VirtualNetwork   string
}
