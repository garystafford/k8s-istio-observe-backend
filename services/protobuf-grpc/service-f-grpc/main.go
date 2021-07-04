// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service F - gRPC/Protobuf

package main

import (
	"bytes"
	"context"
	"encoding/json"
	lrf "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
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
	logLevel     = getEnv("LOG_LEVEL", "info")
	port         = getEnv("PORT", ":50051")
	serviceName  = getEnv("SERVICE_NAME", "Service F")
	message      = getEnv("GREETING", "Hola, from Service F!")
	mongoConn    = getEnv("MONGO_CONN", "mongodb://mongodb:27017/admin")
	queueName    = getEnv("QUEUE_NAME", "service-d.greeting")
	rabbitMQConn = getEnv("RABBITMQ_CONN", "amqp://guest:guest@rabbitmq:5672")
	greetings    []*pb.Greeting
	log         = logrus.New()
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

	collection := client.Database("service-f").Collection("messages")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, greeting)
	if err != nil {
		log.Error(err)
	}
}

func getMessages(rabbitMQConn string) {
	conn, err := amqp.Dial(rabbitMQConn)
	if err != nil {
		log.Error(err)
	}
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			log.Error(err)
		}
	}(conn)

	ch, err := conn.Channel()
	if err != nil {
		log.Error(err)
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Error(err)
		}
	}(ch)

	q, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error(err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"service-f",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error(err)
	}

	forever := make(chan bool)

	go func() {
		for delivery := range msgs {
			var queuedGreeting pb.Greeting
			callMongoDB(deserialize(delivery.Body, &queuedGreeting), mongoConn)
		}
	}()

	<-forever
}

func deserialize(b []byte, greeting *pb.Greeting) *pb.Greeting {
	buf := bytes.NewBuffer(b)
	decoder := json.NewDecoder(buf)
	err := decoder.Decode(&greeting)
	if err != nil {
		log.Error(err)
	}
	log.Debug(greeting)

	return greeting
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
	go getMessages(rabbitMQConn)

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
