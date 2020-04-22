package publicip

type PublicIPInput struct {
	Name             string
	AllocationMethod string
	IPVersion        string
	SKU              string
}
