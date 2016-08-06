package osoc

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/service"
)

func TestMain(m *testing.M) {
	go startServerGRPC()

	// os.Exit() does not respect defer statements
	ret := m.Run()

	stopServerGRPC()
}

var (
	_host    = "172.17.4.50:50051"
	_grpcsvr *grpc.Server
)

func startServerGRPC() {

	lstn, err := net.Listen("tcp", _host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	_grpcsvr = grpc.NewServer()
	//osopb3.RegisterSimpleServiceServer(_grpcsvr, Usrs)
	osopb3.RegisterSimpleManageServiceServer(_grpcsvr, service.Usrs)

	fmt.Printf("grpc server is running on %s\n", _host)

	if err := _grpcsvr.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("quit application\n")

}

func stopServerGRPC() {
	if _grpcsvr != nil {
		time.Sleep(1000)
		_grpcsvr.Stop()
	}
}
