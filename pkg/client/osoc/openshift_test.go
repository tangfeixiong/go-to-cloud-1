package osoc

import (
	"log"
	"testing"

	"github.com/helm/helm-classic/codec"
	projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/client/osoc"
)

var (
	_oso_project string = "tangfeixiong"
)

func TestProject_RestrieveOne(t *testing.T) {
	r, o, err := findOpenshiftProject(context.Background(), _oso_project)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func findOpenshiftProject(context context.Context,
	name string) (*osopb3.ProjectResponseDataArbitrary, *projectapiv1.Project, error) {

	f := osoc.NewIntegrationFactory(_host)

	in := &osopb3.ProjectCreationRequestData{Name: name}
	out, err := f.RetrieveProjectByName(in)
	if err != nil {
		return nil, err
	}
	if p.Raw != nil && len(p.Raw.ObjectBytes) > 0 {
		log.Println(string(out.Raw.ObjectBytes))

		o, err := codec.JSON.Decode(out.Raw.ObjectBytes).One()
		if err != nil {
			log.Printf("Decoder error: %s\n", err)
			return out, nil, nil
		}

		r := new(projectapiv1.Project)
		if err := o.Object(r); err != nil {
			log.Printf("Decode failed: %s\n", err)
			return out, nil, nil
		}
		return out, r, nil
	}
	return out, nil, nil
}
