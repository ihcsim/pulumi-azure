package publicip

import (
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func Up(
	ctx *pulumi.Context,
	cfg *config.Config,
	resourceGroup *core.ResourceGroup,
	tags pulumi.StringMap) (map[string]*network.PublicIp, error) {

	publicIPInput := []*PublicIPInput{}
	if err := cfg.TryObject("publicIP", &publicIPInput); err != nil {
		return nil, err
	}

	publicIPs := map[string]*network.PublicIp{}
	for _, input := range publicIPInput {
		publicIP, err := network.NewPublicIp(ctx, input.Name, &network.PublicIpArgs{
			AllocationMethod:  pulumi.String(input.AllocationMethod),
			IpVersion:         pulumi.String(input.IPVersion),
			Location:          resourceGroup.Location,
			Name:              pulumi.String(input.Name),
			ResourceGroupName: resourceGroup.Name,
			Sku:               pulumi.String(input.SKU),
			Tags:              tags,
		})
		if err != nil {
			return nil, err
		}

		publicIPs[input.Name] = publicIP
	}

	return publicIPs, nil
}

type PublicIPInput struct {
	Name             string
	AllocationMethod string
	IPVersion        string
	SKU              string
}
