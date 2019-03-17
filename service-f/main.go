// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service F

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

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
		log.Error(err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := w.Write([]byte("{\"alive\": true}"))
	if err != nil {
		log.Error(err)
	}
}

func CallMongoDB(trace Trace) {
	log.Info(trace)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONN")))
	if err != nil {
		log.Error(err)
	}

	defer client.Disconnect(nil)

	collection := client.Database("service-f").Collection("messages")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = collection.InsertOne(ctx, trace)
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

func deserialize(b []byte) (t Trace) {
	log.Debug(b)
	var tmpTrace Trace
	buf := bytes.NewBuffer(b)
	decoder := json.NewDecoder(buf)
	err := decoder.Decode(&tmpTrace)
	if err != nil {
		log.Error(err)
	}
	return tmpTrace
}

func init() {
	formatter := runtime.Formatter{ChildFormatter: &log.JSONFormatter{}}
	formatter.Line = true
	log.SetFormatter(&formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	go GetMessages()

	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":80", router))
}
