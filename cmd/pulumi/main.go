package main

import (
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/appsecgroup"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/bastion"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/compute"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/network"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/publicip"
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

			cfg = config.New(ctx, "pulumi-azure")
		)

		resourceGroup, err := resourcegroup.Up(ctx, cfg, commonTags)
		if err != nil {
			return err
		}

		appSecGroups, err := appsecgroup.Up(ctx, cfg, resourceGroup, commonTags)
		if err != nil {
			return err
		}

		virtualNetworks, err := network.Up(ctx, cfg, appSecGroups, resourceGroup, commonTags)
		if err != nil {
			return err
		}

		if _, err := compute.Up(ctx, cfg, appSecGroups, resourceGroup, virtualNetworks, commonTags); err != nil {
			return err
		}

		publicIPs, err := publicip.Up(ctx, cfg, resourceGroup, commonTags)
		if err != nil {
			return err
		}

		if _, err := bastion.Up(ctx, cfg, publicIPs, resourceGroup, virtualNetworks, commonTags); err != nil {
			return err
		}

		return nil
	})
}
