// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service D - gRPC/Protobuf

package main

import (
	"context"
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"

	pb "github.com/garystafford/pb-greeting"
)

var (
	listenerPort = ":" + getEnv("SRV_D_PORT", "50051")
	rabbitConn   = getEnv("RABBITMQ_CONN", "")
	queueName    = getEnv("SRV_D_QUEUE", "service-d")
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

	conn, err := amqp.Dial(rabbitConn)
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
