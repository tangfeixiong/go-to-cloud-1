
package builder

import (
    "fmt"
    "testing"
    
	"github.com/openshift/origin/pkg/version"
)

func TestOVersion(t *testing.T) {
    thisVersion := version.Get().String()
    fmt.Println(thisVersion)
}