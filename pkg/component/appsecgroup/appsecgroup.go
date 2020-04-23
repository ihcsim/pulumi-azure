package appsecgroup

import (
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/network"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func Reconcile(
	ctx *pulumi.Context,
	cfg *config.Config,
	resourceGroup *core.ResourceGroup,
	tags pulumi.StringMap) (map[string]*network.ApplicationSecurityGroup, error) {

	appSecGroupsInput := []*ApplicationSecurityGroupInput{}
	if err := cfg.TryObject("appSecurityGroups", &appSecGroupsInput); err != nil {
		return nil, err
	}

	appSecGroups := map[string]*network.ApplicationSecurityGroup{}
	for _, input := range appSecGroupsInput {
		appSecGroup, err := network.NewApplicationSecurityGroup(ctx, input.Name,
			&network.ApplicationSecurityGroupArgs{
				Location:          resourceGroup.Location,
				Name:              pulumi.String(input.Name),
				ResourceGroupName: resourceGroup.Name,
				Tags:              tags,
			})
		if err != nil {
			return nil, err
		}

		appSecGroups[input.Name] = appSecGroup
	}

	return appSecGroups, nil
}
