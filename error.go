package pulumiazure

import (
	"fmt"
)

type MissingConfigErr struct {
	Name string
	Kind string
}

func (e MissingConfigErr) Error() string {
	return fmt.Sprintf("missing config. name: %s, kind: %s", e.Name, e.Kind)
}
