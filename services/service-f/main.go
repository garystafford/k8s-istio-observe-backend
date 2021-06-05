// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service F
// date: 2021-06-05

package main

import (
	"bytes"
	"context"
	"encoding/json"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	logLevel     = getEnv("LOG_LEVEL", "debug")
	port         = getEnv("PORT", ":8080")
	message      = getEnv("GREETING", "Hola, from Service F!")
	queueName    = getEnv("QUEUE_NAME", "service-d.greeting")
	mongoConn    = getEnv("MONGO_CONN", "mongodb://mongodb:27017/admin")
	rabbitMQConn = getEnv("RABBITMQ_CONN", "amqp://guest:guest@rabbitmq:5672")
)

type Greeting struct {
	ID          string    `json:"id,omitempty"`
	ServiceName string    `json:"service,omitempty"`
	Message     string    `json:"message,omitempty"`
	CreatedAt   time.Time `json:"created,omitempty"`
	Hostname    string    `json:"hostname,omitempty"`
}

var greetings []Greeting

// *** HANDLERS ***

func GreetingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	greetings = nil

	tmpGreeting := Greeting{
		ID:          uuid.New().String(),
		ServiceName: "Service F",
		Message:     message,
		CreatedAt:   time.Now().Local(),
		Hostname:    getHostname(),
	}

	greetings = append(greetings, tmpGreeting)

	callMongoDB(tmpGreeting, mongoConn)

	err := json.NewEncoder(w).Encode(greetings)
	if err != nil {
		log.Error(err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("{\"alive\": true}"))
	if err != nil {
		log.Error(err)
	}
}

// *** UTILITY FUNCTIONS ***

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Error(err)
	}
	return hostname
}

func callMongoDB(greeting Greeting, mongoConn string) {
	log.Info(greeting)
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
			log.Debug(delivery)
			callMongoDB(deserialize(delivery.Body), mongoConn)
		}
	}()

	<-forever
}

func deserialize(b []byte) (t Greeting) {
	log.Debug(b)
	var tmpGreeting Greeting
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

func run() error {
	go getMessages(rabbitMQConn)

	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/greeting", GreetingHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	api.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(port, router)
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
