// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service D - gRPC/Protobuf

package main

import (
	"context"
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	pb "github.com/garystafford/protobuf/greeting/v3"
	"github.com/google/uuid"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	logLevel     = getEnv("LOG_LEVEL", "info")
	port         = getEnv("PORT", ":50051")
	serviceName  = getEnv("SERVICE_NAME", "Service D")
	message      = getEnv("GREETING", "Shalom (שָׁלוֹם), from Service D!")
	queueName    = getEnv("QUEUE_NAME", "service-d.greeting")
	rabbitMQConn = getEnv("RABBITMQ_CONN", "amqp://guest:guest@rabbitmq:5672")
	greetings    []*pb.Greeting
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

	body, err := json.Marshal(&requestGreeting)
	if err != nil {
		log.Error(err)
	}
	sendMessage(body, rabbitMQConn)

	return &pb.GreetingResponse{
		Greeting: greetings,
	}, nil
}

func sendMessage(body []byte, rabbitMQConn string) {
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

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
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
