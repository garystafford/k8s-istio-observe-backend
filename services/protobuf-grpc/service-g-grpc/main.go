// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service G - gRPC/Protobuf

package main

import (
	"context"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"time"

	pb "github.com/garystafford/protobuf/greeting/v3"
)

var (
	logLevel    = getEnv("LOG_LEVEL", "info")
	port        = getEnv("PORT", ":50051")
	serviceName = getEnv("SERVICE_NAME", "Service G")
	message     = getEnv("GREETING", "Ahlan (أهلا), from Service G!")
	mongoConn   = getEnv("MONGO_CONN", "mongodb://mongodb:27017/admin")
	greetings   []*pb.Greeting
)

type greetingServiceServer struct {
	pb.UnimplementedGreetingServiceServer
}

func (s *greetingServiceServer) Greeting(_ context.Context, req *pb.GreetingRequest) (*pb.GreetingResponse, error) {
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

	callMongoDB(req.GetGreeting(), mongoConn)

	return &pb.GreetingResponse{
		Greeting: greetings,
	}, nil
}

func callMongoDB(greeting *pb.Greeting, mongoConn string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConn))
	if err != nil {
		log.Error(err)
	}

	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Error(err)
		}
	}(client, nil)

	collection := client.Database("service-g").Collection("greetings")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, greeting)
	if err != nil {
		log.Error(err)
	}
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
