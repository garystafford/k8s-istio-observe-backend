// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service D
// date: 2021-06-05

package main

import (
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
)

var (
	logLevel     = getEnv("LOG_LEVEL", "debug")
	port         = getEnv("PORT", ":8080")
	serviceName  = getEnv("SERVICE_NAME", "Service D")
	message      = getEnv("GREETING", "Shalom (שָׁלוֹם), from Service D!")
	queueName    = getEnv("QUEUE_NAME", "service-d.greeting")
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

func GreetingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	greetings = nil

	tmpGreeting := Greeting{
		ID:          uuid.New().String(),
		ServiceName: serviceName,
		Message:     message,
		CreatedAt:   time.Now().Local(),
		Hostname:    getHostname(),
	}

	greetings = append(greetings, tmpGreeting)

	err := json.NewEncoder(w).Encode(greetings)
	if err != nil {
		log.Error(err)
	}

	// Headers must be passed for Jaeger Distributed Tracing
	incomingHeaders := []string{
		"x-b3-flags",
		"x-b3-parentspanid",
		"x-b3-sampled",
		"x-b3-spanid",
		"x-b3-traceid",
		"x-ot-span-context",
		"x-request-id",
	}

	rabbitHeaders := amqp.Table{}

	for _, header := range incomingHeaders {
		if r.Header.Get(header) != "" {
			rabbitHeaders[header] = r.Header.Get(header)
		}
	}

	log.Debug(rabbitHeaders)

	body, err := json.Marshal(tmpGreeting)
	sendMessage(rabbitHeaders, body, rabbitMQConn)
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

func sendMessage(headers amqp.Table, body []byte, rabbitMQConn string) {
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
			Headers:     headers,
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
