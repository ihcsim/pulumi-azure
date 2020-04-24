package mock

import (
	"github.com/pulumi/pulumi/sdk/go/common/resource"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

type Mocks int

func (m Mocks) NewResource(
	typeToken, name string,
	inputs resource.PropertyMap,
	provider, id string) (string, resource.PropertyMap, error) {

	return name + "_id", inputs, nil
}

func (m Mocks) Call(
	token string,
	args resource.PropertyMap,
	provider string) (resource.PropertyMap, error) {

	return args, nil
}

func WithCustomMocks(
	project, stack string,
	config map[string]string,
	mocks pulumi.MockResourceMonitor) pulumi.RunOption {

	return func(info *pulumi.RunInfo) {
		info.Project, info.Stack, info.Mocks = project, stack, mocks
		info.Config = config
	}
}
