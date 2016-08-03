package osoc

import (
	"log"
	"testing"

	"github.com/helm/helm-classic/codec"
	projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift"
)

var (
	_go2cloud1_server string = "localhost:50051"
	_oso_project      string = "default"
)

func TestFindProject(t *testing.T) {
	r, err := findOpenshiftProject(Context.Background(), _oso_project)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func findOpenshiftProject(context context.Context,
	name string) (*projectapiv1.Project, error) {

	conn, err := grpc.Dial(_go2cloud1_server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, err
	}
	defer conn.Close()
	c := openshift.NewSimpleManageServiceClient(conn)

	in := openshift.FindProjectRequest{Name: name}
	out, err := c.FindProject(context, &in)
	if err != nil {
		log.Fatalf("could not request: %v", err)
		return nil, err
	}

	o, err := codec.JSON.Decode(out.Odefv1RawData).One()
	if err != nil {
		log.Errorf("could not create decoder into object: %s", err)
		return nil, err
	}

	r := new(projectapiv1.Project)
	if err := o.Object(r); err != nil {
		log.Errorf("could not create decoder into object: %s", err)
		return nil, err
	}
	return r, nil
}
