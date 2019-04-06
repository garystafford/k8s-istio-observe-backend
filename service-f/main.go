// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service F - gRPC

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
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

var (
	greetings []*pb.Greeting
)

func (s *greetingServiceServer) Greeting(ctx context.Context, req *pb.GreetingRequest) (*pb.GreetingResponse, error) {

	tmpGreeting := pb.Greeting{
		Id:      uuid.New().String(),
		Service: "Service-F",
		Message: "Hola, from Service-F!",
		Created: time.Now().Local().String(),
	}

	CallMongoDB(tmpGreeting)

	greetings = append(greetings, &tmpGreeting)

	return &pb.GreetingResponse{
		Greeting: greetings,
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

	collection := client.Database("service-f").Collection("messages")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = collection.InsertOne(ctx, greeting)
	if err != nil {
		log.Error(err)
	}
}

func GetMessages() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONN"))
	if err != nil {
		log.Error(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Error(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"service-d",
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
			log.Debug(delivery)
			CallMongoDB(deserialize(delivery.Body))
		}
	}()

	<-forever
}

func deserialize(b []byte) (t pb.Greeting) {
	log.Debug(b)
	var tmpGreeting pb.Greeting
	buf := bytes.NewBuffer(b)
	decoder := json.NewDecoder(buf)
	err := decoder.Decode(&tmpGreeting)
	if err != nil {
		log.Error(err)
	}
	return tmpGreeting
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
	go GetMessages()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreetingServiceServer(s, &greetingServiceServer{})
	log.Fatal(s.Serve(lis))
}
