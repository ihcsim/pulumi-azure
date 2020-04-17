package resourcegroup

import (
	"github.com/ihcsim/pulumi-azure/v2/config"
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func Up(ctx *pulumi.Context, tags pulumi.StringMap) (*core.ResourceGroup, error) {
	return core.NewResourceGroup(ctx, string(config.ResourceGroup.Name),
		&core.ResourceGroupArgs{
			Location: config.ResourceGroup.Location,
			Tags:     tags,
		})
}
