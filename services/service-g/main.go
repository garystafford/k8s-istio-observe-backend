// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service G

package main

import (
	"context"
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	greetings = nil

	tmpGreeting := Greeting{
		ID:          uuid.New().String(),
		ServiceName: "Service-G",
		Message:     "Ahlan, from Service-G!",
		CreatedAt:   time.Now().Local(),
	}

	greetings = append(greetings, tmpGreeting)

	CallMongoDB(tmpGreeting)

	err := json.NewEncoder(w).Encode(greetings)
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

func CallMongoDB(greeting Greeting) {
	log.Info(greeting)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONN")))
	if err != nil {
		log.Error(err)
	}

	defer client.Disconnect(nil)

	collection := client.Database("service-g").Collection("greetings")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = collection.InsertOne(ctx, greeting)
	if err != nil {
		log.Error(err)
	}
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
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":80", router))
}
