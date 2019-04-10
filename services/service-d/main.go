// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service D

package main

import (
	"context"
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"

	pb "github.com/garystafford/pb-greeting"
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
	greetings = nil

	tmpGreeting := pb.Greeting{
		Id:      uuid.New().String(),
		Service: "Service-D",
		Message: "Shalom, from Service-D!",
		Created: time.Now().Local().String(),
	}

	greetings = append(greetings, &tmpGreeting)

	b, err := json.Marshal(tmpGreeting)
	SendMessage(b)
	if err != nil {
		log.Error(err)
	}

	return &pb.GreetingResponse{
		Greeting: greetings,
	}, nil
}

func SendMessage(b []byte) {
	log.Info(b)

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

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		})
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
	log.Fatal(s.Serve(lis))
}
