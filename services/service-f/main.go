// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service F
// date: 2021-05-24

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os"
	"time"
)

type Greeting struct {
	ID          string    `json:"id,omitempty"`
	ServiceName string    `json:"service,omitempty"`
	Message     string    `json:"message,omitempty"`
	CreatedAt   time.Time `json:"created,omitempty"`
}

var greetings []Greeting

func GreetingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	greetings = nil

	tmpGreeting := Greeting{
		ID:          uuid.New().String(),
		ServiceName: "Service F",
		Message:     "Hola, from Service F!",
		CreatedAt:   time.Now().Local(),
	}

	greetings = append(greetings, tmpGreeting)

	CallMongoDB(tmpGreeting)

	err := json.NewEncoder(w).Encode(greetings)
	if err != nil {
		log.Error(err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := w.Write([]byte("{\"alive\": true}"))
	if err != nil {
		log.Error(err)
	}
}

func CallMongoDB(greeting Greeting) {
	log.Info(greeting)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONN")))
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

func GetMessages() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONN"))
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
		"service-d.greeting",
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

	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/greeting", GreetingHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	api.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":80", router))
}
