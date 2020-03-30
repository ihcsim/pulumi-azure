package main

import (
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

const (
	owner  = "isim-dev"
	region = pulumi.String("WestUS")
)

var (
	vnetCIDR = pulumi.StringArray{
		pulumi.String("10.0.0.0/16"),
	}
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		commonTags := pulumi.StringMap{
			"project": pulumi.String(ctx.Project()),
			"stack":   pulumi.String(ctx.Stack()),
		}

		resourceGroup, err := core.NewResourceGroup(ctx, owner,
			&core.ResourceGroupArgs{
				Location: region,
				Tags:     commonTags,
			})
		if err != nil {
			return err
		}

		network.NewVirtualNetwork(ctx, owner,
			&network.VirtualNetworkArgs{
				AddressSpaces:     vnetCIDR,
				Location:          resourceGroup.Location,
				ResourceGroupName: resourceGroup.Name,
				Tags:              commonTags,
			})

		return nil
	})
}
