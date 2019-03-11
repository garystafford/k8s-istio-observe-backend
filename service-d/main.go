// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service D

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
	ServiceName string    `json:"service,omitempty"`
	Greeting    string    `json:"greeting,omitempty"`
	CreatedAt   time.Time `json:"created,omitempty"`
}

var traces []Trace

func PingHandler(w http.ResponseWriter, r *http.Request) {
	traces = nil

	tmpTrace := Trace{
		ID:          uuid.New().String(),
		ServiceName: "Service-D",
		Greeting:    "Shalom, from Service-D!",
		CreatedAt:   time.Now().Local(),
	}

	traces = append(traces, tmpTrace)
	fmt.Println(traces)

	err := json.NewEncoder(w).Encode(traces)
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.Marshal(tmpTrace)
	SendMessage(b)
	if err != nil {
		log.Fatal(err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := w.Write([]byte("{\"alive\": true}"))
	if err != nil {
		log.Fatal(err)
	}
}

func SendMessage(b []byte) {
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
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		})
	print(b)
	failOnError(err, "Failed to publish a message")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":80", router))
}
