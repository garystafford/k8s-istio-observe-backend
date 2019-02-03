package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mongodb/mongo-go-driver/mongo"
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

	tmpTrace := Trace{ID: uuid.New().String(), ServiceName: "Service-H", CreatedAt: time.Now().Local()}

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
	client, err := mongo.Connect(ctx, os.Getenv("MONGO_CONN"))
	if err != nil {
		panic(err)
	}

	collection := client.Database("service-h").Collection("traces")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = collection.InsertOne(ctx, trace)
	if err != nil {
		panic(err)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/ping", Orchestrator).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}
