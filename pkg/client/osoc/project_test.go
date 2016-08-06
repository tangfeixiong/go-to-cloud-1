package osoc

import (
	"log"
	"testing"

	"github.com/helm/helm-classic/codec"
	projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osobp3"
)

var (
	_oso_project string = "tangfeixiong"
)

func TestProject_Find(t *testing.T) {
	r, err := findOpenshiftProject(context.Background(), _oso_project)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func findOpenshiftProject(context context.Context,
	name string) (*projectapiv1.Project, error) {

	conn, err := grpc.Dial(_host, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, err
	}
	defer conn.Close()
	c := osobp3.NewSimpleManageServiceClient(conn)

	in := osobp3.FindProjectRequest{Name: name}

	out, err := c.FindProject(context, &in)
	if err != nil {
		log.Fatalf("could not request: %v", err)
		return nil, err
	}
	log.Println(string(out.Odefv1RawData))

	o, err := codec.JSON.Decode(out.Odefv1RawData).One()
	if err != nil {
		log.Fatalf("could not create decoder into object: %s", err)
		return nil, err
	}

	r := new(projectapiv1.Project)
	if err := o.Object(r); err != nil {
		log.Fatalf("could not create decoder into object: %s", err)
		return nil, err
	}
	return r, nil
}
