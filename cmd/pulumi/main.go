package main

import (
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/compute"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/network"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/resourcegroup"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		var (
			commonTags = pulumi.StringMap{
				"project": pulumi.String(ctx.Project()),
				"stack":   pulumi.String(ctx.Stack()),
			}

			config = config.New(ctx, "pulumi-azure")
		)

		resourceGroup, err := resourcegroup.Up(ctx, config, commonTags)
		if err != nil {
			return err
		}

		virtualNetworks, err := network.Up(ctx, config, resourceGroup, commonTags)
		if err != nil {
			return err
		}

		if _, err := compute.Up(ctx, config, resourceGroup, virtualNetworks, commonTags); err != nil {
			return err
		}

		return nil
	})
}
