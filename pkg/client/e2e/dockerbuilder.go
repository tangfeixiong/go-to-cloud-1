package e2e

import (
	"fmt"
)

func secretname_for_pull_with_dockerbuilder(buildname string) string {
	return fmt.Sprintf("dockerconfigjson-%s-for", buildname)
}

func secretname_for_push_with_dockerbuilder(buildname string) string {
	return fmt.Sprintf("dockerconfigjson-%s-to", buildname)
}
