// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: gRPC Gateway / Reverse Proxy
// reference: https://github.com/grpc-ecosystem/grpc-gateway

package main

import (
	"flag"
	"github.com/gorilla/handlers"
	"google.golang.org/grpc/metadata"
	"net/http"

	"context"
	gw "github.com/garystafford/pb-greeting"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	prefixTracerState  = "x-b3-"
	zipkinTraceID      = prefixTracerState + "traceid"
	zipkinSpanID       = prefixTracerState + "spanid"
	zipkinParentSpanID = prefixTracerState + "parentspanid"
	zipkinSampled      = prefixTracerState + "sampled"
	zipkinFlags        = prefixTracerState + "flags"
)

var (
	echoEndpoint = flag.String("service-a_endpoint", "service-a:50051", "endpoint of Service-A")
	otHeaders    = []string{
		zipkinTraceID,
		zipkinSpanID,
		zipkinParentSpanID,
		zipkinSampled,
		zipkinFlags}
)

func injectHeadersIntoMetadata(ctx context.Context, req *http.Request) metadata.MD {
	//https://aspenmesh.io/2018/04/tracing-grpc-with-istio/
	var pairs []string
	for _, h := range otHeaders {
		if v := req.Header.Get(h); len(v) > 0 {
			pairs = append(pairs, h, v)
		}
	}
	return metadata.Pairs(pairs...)
}

type annotator func(context.Context, *http.Request) metadata.MD

func chainGrpcAnnotators(annotators ...annotator) annotator {
	return func(c context.Context, r *http.Request) metadata.MD {
		var mds []metadata.MD
		for _, a := range annotators {
			mds = append(mds, a(c, r))
		}
		return metadata.Join(mds...)
	}
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	//mux := runtime.NewServeMux()

	annotators := []annotator{injectHeadersIntoMetadata}

	mux := runtime.NewServeMux(
		runtime.WithMetadata(chainGrpcAnnotators(annotators...)),
	)

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
