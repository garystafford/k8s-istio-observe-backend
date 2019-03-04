package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		ServiceName: "Service-C",
		Greeting: "Konnichiwa, from Service-C!",
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
	//print(os.Getenv("MONGO_CONN"))

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONN")))
	if err != nil {
		panic(err)
	}

	collection := client.Database("service-c").Collection("traces")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = collection.InsertOne(ctx, trace)
	if err != nil {
		panic(err)
	}
}

func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}
