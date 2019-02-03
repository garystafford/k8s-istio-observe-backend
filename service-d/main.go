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
	//time.Sleep(250 * time.Millisecond)

	traces = nil

	tmpTrace := Trace{ID: uuid.New().String(), ServiceName: "Service-D", CreatedAt: time.Now().Local()}

	traces = append(traces, tmpTrace)
	fmt.Println(traces)

	SendMessage(tmpTrace)

	err := json.NewEncoder(w).Encode(traces)
	if err != nil {
		panic(err)
	}
}

func SendMessage(trace Trace) {
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

	err = ch.Publish(
		"",
		q.Name, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(fmt.Sprintf("%v", trace)),
		})
	failOnError(err, "Failed to publish a message")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/ping", Orchestrator).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}
