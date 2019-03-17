// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service D

package main

import (
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
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
		ServiceName: "Service-D",
		Greeting:    "Shalom, from Service-D!",
		CreatedAt:   time.Now().Local(),
	}

	traces = append(traces, tmpTrace)

	err := json.NewEncoder(w).Encode(traces)
	if err != nil {
		log.Error(err)
	}

	b, err := json.Marshal(tmpTrace)
	SendMessage(b)
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

func SendMessage(b []byte) {
	log.Info(b)

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

func init() {
	formatter := runtime.Formatter{ChildFormatter: &log.JSONFormatter{}}
	formatter.Line = true
	log.SetFormatter(&formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":80", router))
}
