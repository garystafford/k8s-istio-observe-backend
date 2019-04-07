// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service B - gRPC

package main

import (
	"context"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"

	pb "../greeting"
)

const (
	port = ":50051"
)

type greetingServiceServer struct {
}

var (
	greetings []*pb.Greeting
)

func (s *greetingServiceServer) Greeting(ctx context.Context, req *pb.GreetingRequest) (*pb.GreetingResponse, error) {
	tmpGreeting := pb.Greeting{
		Id:      uuid.New().String(),
		Service: "Service-B",
		Message: "Namaste, from Service-B!",
		Created: time.Now().Local().String(),
	}

	greetings = append(greetings, &tmpGreeting)

	CallGrpcService("service-d:50051")
	CallGrpcService("service-e:50051")

	return &pb.GreetingResponse{
		Greeting: greetings,
	}, nil
}

func CallGrpcService(address string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreetingServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := pb.GreetingRequest{}
	greeting, err := client.Greeting(ctx, &req)
	log.Info(greeting.GetGreeting())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	for _, greeting := range greeting.GetGreeting() {
		greetings = append(greetings, greeting)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func init() {
	formatter := runtime.Formatter{ChildFormatter: &log.JSONFormatter{}}
	formatter.Line = true
	log.SetFormatter(&formatter)
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(getEnv("LOG_LEVEL", "info"))
	if err != nil {
		log.Error(err)
	}
	log.SetLevel(level)
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreetingServiceServer(s, &greetingServiceServer{})
	log.Fatal(s.Serve(lis))
}
