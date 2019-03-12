// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service G

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
		ServiceName: "Service-G",
		Greeting:    "Ahlan, from Service-G!",
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

	collection := client.Database("service-g").Collection("traces")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = collection.InsertOne(ctx, trace)
	if err != nil {
		log.WithField("func", "collection.InsertOne()").Fatal(err)
	}
}

func init() {
	log.SetFormatter(&joonix.FluentdFormatter{})
}

func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	err := http.ListenAndServe(":80", handler)
	if err != nil {
		log.WithField("func", "http.ListenAndServe()").Fatal(err)
	}
}
