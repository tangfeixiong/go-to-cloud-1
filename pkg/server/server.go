package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/gengo/grpc-gateway/runtime"
	"github.com/golang/glog"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/service"
)

type APIContextServer struct {
	grpcServer  *grpc.Server
	grpcGateway *service.UserResource
	webService  *restful.WebService
}

var (
	APIServer  *APIContextServer
	swaggerDir = "docs/apidocs.json"
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

func runGrpcServer() error {
	lstn, err := net.Listen("tcp", ":8086")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		return err
	}

	fmt.Printf("grpc server is running on %s\n", ":8086")

	//s := grpc.NewServer()
	//examples.RegisterEchoServiceServer(s, newEchoServer())
	//examples.RegisterFlowCombinationServer(s, newFlowCombinationServer())

	//abe := newABitOfEverythingServer()
	//examples.RegisterABitOfEverythingServiceServer(s, abe)
	//examples.RegisterStreamServiceServer(s, abe)

	//s.Serve(l)

	if err := APIServer.grpcServer.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		return err
	}
	return nil
}

// newGateway returns a new gateway server which translates HTTP into gRPC.
func newGateway(ctx context.Context, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	mux := runtime.NewServeMux(opts...)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := openshift.RegisterSimpleManageServiceHandlerFromEndpoint(ctx, mux, "localhost:8087", dialOpts)
	if err != nil {
		return nil, err
	}
	//err = openshift.RegisterStreamServiceHandlerFromEndpoint(ctx, mux, "localhost:8087", dialOpts)
	//if err != nil {
	//	return nil, err
	//}
	//err = examplepb.RegisterABitOfEverythingServiceHandlerFromEndpoint(ctx, mux, *abeEndpoint, dialOpts)
	//if err != nil {
	//	return nil, err
	//}
	//err = examplepb.RegisterFlowCombinationHandlerFromEndpoint(ctx, mux, *flowEndpoint, dialOpts)
	//if err != nil {
	//	return nil, err
	//}
	return mux, nil
}

// allowCORS allows Cross Origin Resoruce Sharing from any origin.
// Don't do this without consideration in production systems.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				func(w http.ResponseWriter, r *http.Request) {
					headers := []string{"Content-Type", "Accept"}
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
					methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
					glog.Infof("preflight request for %s", r.URL.Path)
					return
				}(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

// Run starts a HTTP server and blocks forever if successful.
func runGrpcGateway(address string, opts ...runtime.ServeMuxOption) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
			glog.Errorf("Not Found: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		glog.Infof("Serving %s", r.URL.Path)
		p := strings.TrimPrefix(r.URL.Path, "/swagger/")
		p = path.Join(swaggerDir, p)
		http.ServeFile(w, r, p)
	})

	gw, err := newGateway(ctx, opts...)
	if err != nil {
		return err
	}
	mux.Handle("/", gw)

	service.Usrs.ContextBase = ctx
	service.Usrs.HttpMuxs = []*http.ServeMux{mux}
	service.Usrs.GatewayMux = gw.(*runtime.ServeMux)

	http.ListenAndServe(address, allowCORS(mux))
	return nil
}

func Run() {
	errCh := make(chan error, 2)
	go func() {
		if err := runGrpcServer(); err != nil {
			errCh <- fmt.Errorf("cannot run gRPC service: %v", err)
		}
	}()

	go func() {
		if err := runGrpcGateway(":8080"); err != nil {
			errCh <- fmt.Errorf("cannot run gateway service: %v", err)
		} else {
			APIServer.grpcGateway = service.Usrs
		}
	}()

	ch := make(chan int, 1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		if err := service.Run(); err != nil {
			fmt.Printf("Could not start REST service: %s\n", err)
			ch <- 1
		} else {
			ch <- 0
		}
	}()

	select {
	case err := <-errCh:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	case status := <-ch:
		os.Exit(status)
	}
}
