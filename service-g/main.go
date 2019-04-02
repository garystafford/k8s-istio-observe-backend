// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service G - gRPC

package main

import (
	"context"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (s *greetingServiceServer) Greeting(ctx context.Context, req *pb.GreetingRequest) (*pb.GreetingResponse, error) {

	tmpGreeting := pb.Greeting{
		Id:      uuid.New().String(),
		Service: "Service-G",
		Message: "Ahlan, from Service-G!",
		Created: time.Now().Local().String(),
	}

	CallMongoDB(tmpGreeting)

	return &pb.GreetingResponse{
		Greeting: &tmpGreeting,
	}, nil
}

func (s *greetingServiceServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {

	tmpHealth := pb.Health{
		Alive: "ok",
	}

	return &pb.HealthResponse{
		Health: &tmpHealth,
	}, nil
}

func CallMongoDB(greeting pb.Greeting) {
	log.Info(greeting)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONN")))
	if err != nil {
		log.Error(err)
	}

	defer client.Disconnect(nil)

	collection := client.Database("service-g").Collection("greetings")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = collection.InsertOne(ctx, greeting)
	if err != nil {
		log.Error(err)
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
	s.Serve(lis)
}