package rest

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/turtacn/guocedb/common/log"
	// Import the generated gRPC gateway code.
	// This import will not work until the .proto file is compiled.
	// mgmtv1 "github.com/turtacn/guocedb/api/protobuf/mgmt/v1"
)

// NewGateway creates a new RESTful API gateway that proxies requests to the gRPC management service.
func NewGateway(ctx context.Context, grpcAddr string) (http.Handler, error) {
	mux := runtime.NewServeMux()

	// Options for dialing the gRPC server.
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register the gRPC service handlers with the gateway.
	// This is a placeholder for the generated code. The following line will not compile
	// until the `protoc` command with `protoc-gen-grpc-gateway` is run.
	/*
		err := mgmtv1.RegisterManagementServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
		if err != nil {
			return nil, err
		}
	*/

	log.Infof("REST API gateway configured to proxy to gRPC server at %s", grpcAddr)
	return mux, nil
}

// TODO:
// 1. Install protoc, protoc-gen-go, protoc-gen-go-grpc, and protoc-gen-grpc-gateway.
// 2. Add "google.api.http" annotations to the management.proto file to define HTTP mappings.
//    Example:
//      rpc GetStatus(GetStatusRequest) returns (GetStatusResponse) {
//        option (google.api.http) = {
//          get: "/v1/status"
//        };
//      }
// 3. Run protoc to generate the Go gRPC code and the gRPC-gateway reverse proxy code.
// 4. Uncomment the code above to enable the gateway.
