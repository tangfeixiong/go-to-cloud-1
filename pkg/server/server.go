package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/emicklei/go-restful"

	"github.com/golang/glog"

	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/service"
)

type APIContextServer struct {
	grpcServer *grpc.Server
	webService *restful.WebService
}

var (
	APIServer *APIContextServer
)

func init() {
	APIServer = new(APIContextServer)
	//APIServer.grpcServer = grpc.NewServer()
	//APIServer.webService = new(restful.WebService)
}

func (s *APIContextServer) GRPCServer(server *grpc.Server) *APIContextServer {
	if server == nil {
		log.Fatal(fmt.Errorf("gRPC server not found: %v", s))
	}
	s.grpcServer = server
	openshift.RegisterSimpleServiceServer(server, service.Usrs)
	openshift.RegisterSimpleManageServiceServer(server, service.Usrs)
	return s
}

func (s *APIContextServer) WebService(service *restful.WebService) *APIContextServer {
	s.webService = service
	return s
}

func Run() {
	go service.Run()

	lstn, err := net.Listen("tcp", ":8086")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("grpc server is running on %s\n", "")

	if err := APIServer.grpcServer.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	glog.Info("quit application\n")
}
