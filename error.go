package pulumiazure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

type MissingConfigErr struct {
	Name pulumi.String
	Kind string
}

func (e MissingConfigErr) Error() string {
	return fmt.Sprintf("missing config. name: %s, kind: %s", e.Name, e.Kind)
}
