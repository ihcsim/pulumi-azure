package main

import (
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/compute"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/network"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/resourcegroup"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		commonTags := pulumi.StringMap{
			"project": pulumi.String(ctx.Project()),
			"stack":   pulumi.String(ctx.Stack()),
		}

		resourceGroup, err := resourcegroup.Up(ctx, commonTags)
		if err != nil {
			return err
		}

		virtualNetworks, err := network.Up(ctx, resourceGroup, commonTags)
		if err != nil {
			return err
		}

		if _, err := compute.Up(ctx, resourceGroup, virtualNetworks, commonTags); err != nil {
			return err
		}

		return nil
	})
}
