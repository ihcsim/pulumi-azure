package main

import (
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/appsecgroup"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/bastion"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/compute"
	"github.com/ihcsim/pulumi-azure/v2/pkg/component/loadbalancer"
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

		resourceGroup, err := resourcegroup.Reconcile(ctx, cfg, commonTags)
		if err != nil {
			return err
		}

		appSecGroups, err := appsecgroup.Reconcile(ctx, cfg, resourceGroup, commonTags)
		if err != nil {
			return err
		}

		virtualNetworks, err := network.Reconcile(ctx, cfg, appSecGroups, resourceGroup, commonTags)
		if err != nil {
			return err
		}

		if _, err := compute.Reconcile(ctx, cfg, appSecGroups, resourceGroup, virtualNetworks, commonTags); err != nil {
			return err
		}

		publicIPs, err := publicip.Reconcile(ctx, cfg, resourceGroup, commonTags)
		if err != nil {
			return err
		}

		if _, err := bastion.Reconcile(ctx, cfg, publicIPs, resourceGroup, virtualNetworks, commonTags); err != nil {
			return err
		}

		if _, err := loadbalancer.Reconcile(ctx, cfg, publicIPs, resourceGroup, virtualNetworks, commonTags); err != nil {
			return err
		}

		return nil
	})
}
