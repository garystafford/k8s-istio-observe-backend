// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service F

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	joonix "github.com/joonix/log"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type Trace struct {
	ID          string    `json:"id,omitempty"`
	ServiceName string    `json:"service,omitempty"`
	Greeting    string    `json:"greeting,omitempty"`
	CreatedAt   time.Time `json:"created,omitempty"`
}

var traces []Trace

func PingHandler(w http.ResponseWriter, r *http.Request) {
	traces = nil

	tmpTrace := Trace{
		ID:          uuid.New().String(),
		ServiceName: "Service-F",
		Greeting:    "Hola, from Service-F!",
		CreatedAt:   time.Now().Local(),
	}

	traces = append(traces, tmpTrace)

	CallMongoDB(tmpTrace)

	err := json.NewEncoder(w).Encode(traces)
	if err != nil {
		log.WithField("func", "json.NewEncoder()").Fatal(err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := w.Write([]byte("{\"alive\": true}"))
	if err != nil {
		log.WithField("func", "w.Write()").Fatal(err)
	}
}

func CallMongoDB(trace Trace) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONN")))
	if err != nil {
		log.WithField("func", "mongo.Connect()").Fatal(err)
	}

	defer client.Disconnect(nil)

	collection := client.Database("service-f").Collection("messages")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = collection.InsertOne(ctx, trace)
	if err != nil {
		log.WithField("func", "collection.InsertOne()").Fatal(err)
	}
}

func GetMessages() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONN"))
	if err != nil {
		log.WithField("func", "amqp.Dial()").Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.WithField("func", "conn.Channel()").Fatal(err)
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
		log.WithField("func", "ch.QueueDeclare()").Fatal(err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.WithField("func", "ch.Consume()").Fatal(err)
	}

	forever := make(chan bool)

	go func() {
		for delivery := range msgs {
			log.WithField("func", "GetMessages()").Infof("message: %s", delivery)
			CallMongoDB(deserialize(delivery.Body))
		}
	}()

	log.Infof(" [*] Waiting for messages...")
	<-forever
}

func deserialize(b []byte) (t Trace) {
	var tmpTrace Trace
	log.WithField("func", "amqp.Publishing()").Infof("body: %s", b)
	buf := bytes.NewBuffer(b)
	decoder := json.NewDecoder(buf)
	err := decoder.Decode(&tmpTrace)
	if err != nil {
		log.WithField("func", "decoder.Decode()").Fatal(err)
	}
	return tmpTrace
}

func init() {
	log.SetFormatter(&joonix.FluentdFormatter{})
}

func main() {
	GetMessages()
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	err := http.ListenAndServe(":80", handler)
	if err != nil {
		log.WithField("func", "http.ListenAndServe()").Fatal(err)
	}
}
