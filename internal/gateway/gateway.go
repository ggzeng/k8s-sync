package gateway

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"k8sync/gen/proto/k8sync/v1"
	"k8sync/internal/config"
	"k8sync/internal/process"
	log "k8sync/pkg/logger"
	"k8sync/third_party"
)

// getOpenAPIHandler serves an OpenAPI UI.
func getOpenAPIHandler() http.Handler {
	err := mime.AddExtensionType(".svg", "image/svg+xml")
	if err != nil {
		return nil
	}
	// Use subdirectory in embedded files
	subFS, err := fs.Sub(third_party.OpenAPI, "OpenAPI")
	if err != nil {
		panic("couldn't create sub filesystem: " + err.Error())
	}
	return http.StripPrefix("/swagger", http.FileServer(http.FS(subFS)))
}

// Start runs the gRPC-Gateway, dialling the provided address.
func Start(ctx context.Context) error {
	grpclog.SetLoggerV2(log.GetGrpcLogger())

	grpcAddr := config.GetAppGrpcDomain()
	startGrpcServer(ctx, grpcAddr)
	return startHttpServer(ctx, grpcAddr)
}

func startGrpcServer(ctx context.Context, grpcAddr string) {
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
	gsv := grpc.NewServer()
	pb.RegisterHealthServiceServer(gsv, process.NewHealth())

	// Serve gRPC Server
	log.Info("Serving gRPC on http://", grpcAddr)
	go func() {
		defer gsv.GracefulStop()
		<-ctx.Done()
	}()
	go func() {
		if err := gsv.Serve(lis); err != nil {
			log.Fatalf("grpc server shutdown: %s", err)
		}
	}()
}

func startHttpServer(ctx context.Context, grpcAddr string) error {
	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	conn, err := grpc.DialContext(
		ctx,
		"dns:///"+grpcAddr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}

	gwmux := runtime.NewServeMux()
	err = pb.RegisterHealthServiceHandler(ctx, gwmux, conn)
	if err != nil {
		return fmt.Errorf("register user service handler failed: %w", err)
	}

	swagger := getOpenAPIHandler()
	gatewayAddr := config.GetAppHttpDomain()
	gwServer := &http.Server{
		Addr: gatewayAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				swagger.ServeHTTP(w, r)
				return
			}
			gwmux.ServeHTTP(w, r)
		}),
	}

	log.Info("serving gRPC-Gateway and OpenAPI Documentation on http://", gatewayAddr)
	go func() {
		<-ctx.Done()
		log.Infof("shutting down http gateway server")
		if err = gwServer.Shutdown(context.Background()); err != nil {
			log.Errorf("failed to shutdown http gateway server: %s", err)
		}
	}()
	go func() {
		log.Warnf("serving grpc-gateway server %s", gwServer.ListenAndServe())
	}()

	return nil
}
