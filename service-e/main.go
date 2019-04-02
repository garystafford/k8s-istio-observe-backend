// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service E

package main

import (
	"context"
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	pb "../greeting"
)

type Greeting struct {
	ID          string    `json:"id,omitempty"`
	ServiceName string    `json:"service,omitempty"`
	Message     string    `json:"message,omitempty"`
	CreatedAt   time.Time `json:"created,omitempty"`
}

var greetings []pb.Greeting

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	log.Debug(r)

	greetings = nil

	//CallGrpcService("service-g:50051")
	//CallGrpcService("service-h:50051")

	CallGrpcService("localhost:50051")
	CallGrpcService("localhost:50052")

	tmpGreeting := pb.Greeting{
		Id:      uuid.New().String(),
		Service: "Service-E",
		Message: "Bonjour, de Service-E!",
		Created: time.Now().Local().String(),
	}

	greetings = append(greetings, tmpGreeting)

	err := json.NewEncoder(w).Encode(greetings)
	if err != nil {
		log.Error(err)
	}
}

func CallGrpcService(address string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreetingServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := pb.GreetingRequest{}
	greeting, err := client.Greeting(ctx, &req)
	log.Info(greeting.GetGreeting())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	greetings = append(greetings, *greeting.GetGreeting())
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := w.Write([]byte("{\"alive\": true}"))
	if err != nil {
		log.Error(err)
	}
}

func CallNextServiceWithTrace(url string, w http.ResponseWriter, r *http.Request) {
	log.Info(url)

	var tmpGreetings []pb.Greeting

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
	}

	headers := []string{
		"x-request-id",
		"x-b3-traceid",
		"x-b3-spanid",
		"x-b3-parentspanid",
		"x-b3-sampled",
		"x-b3-flags",
		"x-ot-span-context",
	}

	for _, header := range headers {
		if r.Header.Get(header) != "" {
			req.Header.Add(header, r.Header.Get(header))
		}
	}

	log.Info(req)

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		log.Error(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
	}

	err = json.Unmarshal(body, &tmpGreetings)
	if err != nil {
		log.Error(err)
	}

	for _, r := range tmpGreetings {
		greetings = append(greetings, r)
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
	log.Fatal(http.ListenAndServe(":8081", router))
}
