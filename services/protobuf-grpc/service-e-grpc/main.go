// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service E - gRPC/Protobuf

package main

import (
	"context"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	ot "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"net/http"
	"os"
	"time"

	pb "github.com/garystafford/protobuf/greeting/v3"
)

var (
	logLevel    = getEnv("LOG_LEVEL", "info")
	port        = getEnv("PORT", ":50051")
	serviceName = getEnv("SERVICE_NAME", "Service E")
	message     = getEnv("GREETING", "Bonjour, from Service E!")
	URLServiceG = getEnv("SERVICE_G_URL", "service-g:50051")
	URLServiceH = getEnv("SERVICE_H_URL", "service-h:50051")
	greetings   []*pb.Greeting
)

type greetingServiceServer struct {
	pb.UnimplementedGreetingServiceServer
}

func (s *greetingServiceServer) Greeting(ctx context.Context, _ *pb.GreetingRequest) (*pb.GreetingResponse, error) {
	greetings = nil

	requestGreeting := pb.Greeting{
		Id:       uuid.New().String(),
		Service:  serviceName,
		Message:  message,
		Created:  time.Now().Local().String(),
		Hostname: getHostname(),
	}

	greetings = append(greetings, &requestGreeting)

	callGrpcService(ctx, &requestGreeting, URLServiceG)
	callGrpcService(ctx, &requestGreeting, URLServiceH)

	return &pb.GreetingResponse{
		Greeting: greetings,
	}, nil
}

func callGrpcService(ctx context.Context, requestGreeting *pb.Greeting, address string) {
	conn, err := createGRPCConn(ctx, address)
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Error(err)
		}
	}(conn)

	headersIn, _ := metadata.FromIncomingContext(ctx)
	log.Debugf("headersIn: %s", headersIn)

	client := pb.NewGreetingServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	ctx = metadata.NewOutgoingContext(context.Background(), headersIn)

	headersOut, _ := metadata.FromOutgoingContext(ctx)
	log.Debugf("headersOut: %s", headersOut)

	defer cancel()

	responseGreetings, err := client.Greeting(ctx, &pb.GreetingRequest{Greeting: requestGreeting})
	if err != nil {
		log.Fatal(err)
	}
	log.Info(responseGreetings.GetGreeting())

	for _, responseGreeting := range responseGreetings.GetGreeting() {
		greetings = append(greetings, responseGreeting)
	}
}

func createGRPCConn(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithUnaryInterceptor(grpcprometheus.UnaryClientInterceptor),
		grpc.WithUnaryInterceptor(
			opentracing.UnaryClientInterceptor(
				opentracing.WithTracer(ot.GlobalTracer()))),
		grpc.WithInsecure(),
		grpc.WithBlock())
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return conn, nil
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Error(err)
	}
	return hostname
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func run() error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcprometheus.UnaryServerInterceptor),
	)
	pb.RegisterGreetingServiceServer(grpcServer, &greetingServiceServer{})
	grpcprometheus.Register(grpcServer)
	http.Handle("/metrics", promhttp.Handler())

	return grpcServer.Serve(lis)
}

func init() {
	formatter := runtime.Formatter{ChildFormatter: &log.JSONFormatter{}}
	formatter.Line = true
	log.SetFormatter(&formatter)
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Error(err)
	}
	log.SetLevel(level)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
