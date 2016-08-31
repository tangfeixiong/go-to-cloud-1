package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	//"os/signal"
	"path"
	"strings"
	"sync"
	//"syscall"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/gengo/grpc-gateway/runtime"
	"github.com/golang/glog"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/service"
)

type AppContext struct {
	grpcServer  *grpc.Server
	grpcGateway *service.UserResource
	webService  *restful.WebService
}

var (
	AppServer  *AppContext
	swaggerDir = "docs/apidocs.json"
	wg         sync.WaitGroup
)

func init() {
	AppServer = new(AppContext)
	//AppServer.grpcServer = grpc.NewServer()
	//AppServer.webService = new(restful.WebService)
}

func (s *AppContext) GRPCServer(server *grpc.Server) *AppContext {
	if server == nil {
		log.Fatal(fmt.Errorf("gRPC server not found: %v", s))
	}
	s.grpcServer = server
	osopb3.RegisterSimpleServiceServer(server, service.Usrs)
	//osopb3.RegisterSimpleManageServiceServer(server, service.Usrs)
	return s
}

func (s *AppContext) WebService(service *restful.WebService) *AppContext {
	s.webService = service
	return s
}

func runGrpcServer() error {
	host := ":50051"
	if v, ok := os.LookupEnv("APAAS_GRPC_PORT"); ok && v != "" {
		host = v
	}

	fmt.Printf("Listenning tcp on %s\n", host)

	lstn, err := net.Listen("tcp", host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gRPC Server died: %s\n", err)
		return err
	}

	fmt.Println("Starting grpc server")

	//s := grpc.NewServer()
	//examples.RegisterEchoServiceServer(s, newEchoServer())
	//examples.RegisterFlowCombinationServer(s, newFlowCombinationServer())

	//abe := newABitOfEverythingServer()
	//examples.RegisterABitOfEverythingServiceServer(s, abe)
	//examples.RegisterStreamServiceServer(s, abe)

	//s.Serve(l)

	if err := AppServer.grpcServer.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "gRPC Server died: %s\n", err)
		//_ = lstn.Close()
		return err
	}
	return nil
}

// newGateway returns a new gateway server which translates HTTP into gRPC.
func newGateway(ctx context.Context, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	mux := runtime.NewServeMux(opts...)
	//dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	//err := osopb3.RegisterSimpleManageServiceHandlerFromEndpoint(ctx, mux, "localhost:8086", dialOpts)
	//if err != nil {
	//	return nil, err
	//}
	//err = osopb3.RegisterStreamServiceHandlerFromEndpoint(ctx, mux, *abeEndpoint, dialOpts)
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

	fmt.Printf("Starting grpc gateway on %s\n", address)

	return http.ListenAndServe(address, allowCORS(mux))
}

func Run() {
	go func() {
		wg.Add(1)
		defer wg.Done()
		if err := runGrpcServer(); err != nil {
			AppServer.grpcServer = nil
			log.Fatal(err)
		}
	}()

	time.Sleep(1000 * time.Millisecond)
	if AppServer.grpcServer == nil {
		os.Exit(1)
	}

	go func() {
		fmt.Println("Starting dispatcher")
		service.Usrs.Dispatch(&wg)
	}()
	/*
		//errCh := make(chan error, 2)
		go func() {
			time.Sleep(1000 * time.Millisecond)
			if AppServer.grpcServer == nil {
				wg.Done()
				return
			}
			AppServer.grpcGateway = service.Usrs
			if err := runGrpcGateway(":50052"); err != nil {
				errCh <- fmt.Errorf("cannot run gateway service: %v", err)
				wg.Done()
			}
		}()

		go func() {
			wg.Wait()
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
		default:
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

			// Block until a signal is received.
			<-c
		}*/
	wg.Wait()
	fmt.Println("Stopped.")
}
