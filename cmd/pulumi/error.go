package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

type missingConfigErr struct {
	name pulumi.String
	kind string
}

func (e missingConfigErr) Error() string {
	return fmt.Sprintf("missing config. name: %s, kind: %s", e.name, e.kind)
}
