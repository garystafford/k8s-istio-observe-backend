// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: gRPC Gateway / Reverse Proxy
// reference: https://github.com/grpc-ecosystem/grpc-gateway

package main

import (
	"flag"
	"github.com/gorilla/handlers"
	"net/http"

	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	gw "github.com/garystafford/pb-greeting"
)

var (
	echoEndpoint = flag.String("service-a_endpoint", "service-a:50051", "endpoint of Service-A")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	// https://qiita.com/ushio_s/items/a442fa53a8a31b87a360
	newMux := handlers.CORS(
		handlers.AllowedMethods([]string{"GET", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}),
	)(mux)

	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterGreetingServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(":80", newMux)
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
