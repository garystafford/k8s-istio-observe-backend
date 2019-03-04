package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/streadway/amqp"
	"log"
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
		ID: uuid.New().String(),
		ServiceName: "Service-F",
		Greeting: "Hola, from Service-F!",
		CreatedAt: time.Now().Local(),
	}

	traces = append(traces, tmpTrace)
	fmt.Println(traces)

	CallMongoDB(tmpTrace)

	err := json.NewEncoder(w).Encode(traces)
	if err != nil {
		panic(err)
	}
}

func CallMongoDB(trace Trace) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONN")))
	if err != nil {
		panic(err)
	}

	collection := client.Database("service-f").Collection("messages")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = collection.InsertOne(ctx, trace)
	if err != nil {
		panic(err)
	}
}

func GetMessages() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONN"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"service-d",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for delivery := range msgs {
			log.Printf("Received a message: %s", delivery)
			CallMongoDB(deserialize(delivery.Body))
		}
	}()

	log.Printf(" [*] Waiting for messages...")
	<-forever
}

func deserialize(b []byte) (t Trace) {
	var tmpTrace Trace
	log.Printf("Body: %s", b)
	buf := bytes.NewBuffer(b)
	decoder := json.NewDecoder(buf)
	err := decoder.Decode(&tmpTrace)
	if err != nil {
		panic(err)
	}
	return tmpTrace
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	GetMessages()
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}
