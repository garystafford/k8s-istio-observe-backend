// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service B - gRPC/Protobuf

package main

import (
	"context"
	lrf "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"os"
	"time"

	pb "github.com/garystafford/protobuf/greeting/v3"
)

var (
	logLevel    = getEnv("LOG_LEVEL", "info")
	port        = getEnv("PORT", ":50051")
	serviceName = getEnv("SERVICE_NAME", "Service B")
	message     = getEnv("GREETING", "Namasté (नमस्ते), from Service B!")
	URLServiceD = getEnv("SERVICE_D_URL", "service-d:50051")
	URLServiceE = getEnv("SERVICE_E_URL", "service-e:50051")
	greetings   []*pb.Greeting
	log         = logrus.New()
)

type greetingServiceServer struct {
	pb.UnimplementedGreetingServiceServer
}

func (s *greetingServiceServer) Greeting(ctx context.Context, req *pb.GreetingRequest) (*pb.GreetingResponse, error) {
	log.Debugf("GreetingRequest: %v", req.GetGreeting())

	greetings = nil

	requestGreeting := pb.Greeting{
		Id:       uuid.New().String(),
		Service:  serviceName,
		Message:  message,
		Created:  time.Now().Local().String(),
		Hostname: getHostname(),
	}

	greetings = append(greetings, &requestGreeting)

	callGrpcService(ctx, &requestGreeting, URLServiceD)
	callGrpcService(ctx, &requestGreeting, URLServiceE)

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
	grpcServer := grpc.NewServer()
	pb.RegisterGreetingServiceServer(grpcServer, &greetingServiceServer{})
	return grpcServer.Serve(lis)
}

func init() {
	childFormatter := logrus.JSONFormatter{}
	runtimeFormatter := &lrf.Formatter{ChildFormatter: &childFormatter}
	runtimeFormatter.Line = true
	log.Formatter = runtimeFormatter
	log.Out = os.Stdout
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Error(err)
	}
	log.Level = level
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
