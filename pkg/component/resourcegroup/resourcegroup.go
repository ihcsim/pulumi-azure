package resourcegroup

import (
	"github.com/pulumi/pulumi-azure/sdk/go/azure/core"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func Up(ctx *pulumi.Context, cfg *config.Config, tags pulumi.StringMap) (*core.ResourceGroup, error) {
	var input ResourceGroupInput
	if err := cfg.TryObject("resourceGroup", &input); err != nil {
		return nil, err
	}

	return core.NewResourceGroup(ctx, string(input.Name),
		&core.ResourceGroupArgs{
			Location: pulumi.String(input.Location),
			Tags:     tags,
		})
}
