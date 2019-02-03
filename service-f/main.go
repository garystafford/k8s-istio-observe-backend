package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"os"
	"time"
)

type Trace struct {
	ID          string    `json:"id,omitempty"`
	ServiceName string    `json:"serviceName,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
}

var traces []Trace

func Orchestrator(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(1000 * time.Millisecond)

	traces = nil

	tmpTrace := Trace{ID: uuid.New().String(), ServiceName: "Service-F", CreatedAt: time.Now().Local()}

	traces = append(traces, tmpTrace)
	fmt.Println(traces)

	err := json.NewEncoder(w).Encode(traces)
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
		for d := range msgs {
			log.Printf("Received a message: %s", d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	GetMessages()
	router := mux.NewRouter()
	router.HandleFunc("/ping", Orchestrator).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}
