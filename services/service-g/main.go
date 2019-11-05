// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service G - gRPC/Protobuf

package main

import (
	"context"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"

	pb "github.com/garystafford/pb-greeting"
)

var (
	listenerPort = ":" + getEnv("SRV_G_PORT", "50051")
	mongoConn    = getEnv("MONGO_CONN", "")
	dbName       = getEnv("SRV_G_DB", "service-g")
	logLevel     = getEnv("LOG_LEVEL", "info")
)

type greetingServiceServer struct {
}

var (
	greetings []*pb.Greeting
)

func (s *greetingServiceServer) Greeting(ctx context.Context, req *pb.GreetingRequest) (*pb.GreetingResponse, error) {
	greetings = nil

	tmpGreeting := pb.Greeting{
		Id:      ksuid.New().String(),
		Service: "Service-G",
		Message: "Ahlan, from Service-G!",
		Created: time.Now().Local().String(),
	}

	greetings = append(greetings, &tmpGreeting)

	CallMongoDB(tmpGreeting)

	return &pb.GreetingResponse{
		Greeting: greetings,
	}, nil
}

func CallMongoDB(greeting pb.Greeting) {
	log.Info(greeting)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConn))
	if err != nil {
		log.Error(err)
	}

	defer client.Disconnect(nil)

	collection := client.Database(dbName).Collection("greetings")
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
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Error(err)
	}
	log.SetLevel(level)
}

func main() {
	lis, err := net.Listen("tcp", listenerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreetingServiceServer(s, &greetingServiceServer{})
	log.Fatal(s.Serve(lis))
}
