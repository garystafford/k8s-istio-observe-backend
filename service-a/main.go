// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service A

package main

import (
	"encoding/json"
	"github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
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
	log.Info(r)

	traces = nil
	CallNextService("http://service-b/api/ping")
	CallNextService("http://service-c/api/ping")

	tmpTrace := Trace{
		ID:          uuid.New().String(),
		ServiceName: "Service-A",
		Greeting:    "Hello, from Service-A!",
		CreatedAt:   time.Now().Local(),
	}

	traces = append(traces, tmpTrace)

	err := json.NewEncoder(w).Encode(traces)
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


func ResponseStatusHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	statusCode, err := strconv.Atoi(params["code"])
	if err != nil {
		log.Error(err)
	}
	w.WriteHeader(statusCode)

}

func CallNextService(url string) {
	log.Info(url)
	var tmpTraces []Trace
	response, err := http.Get(url)
	if err != nil {
		log.Error(err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		err := json.Unmarshal(data, &tmpTraces)
		if err != nil {
			log.Error(err)
		}

		for _, r := range tmpTraces {
			traces = append(traces, r)
		}
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
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
	})
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/status/{code}", ResponseStatusHandler).Methods("GET", "OPTIONS")
	api.Handle("/metrics", promhttp.Handler())
	handler := c.Handler(router)
	log.Fatal(http.ListenAndServe(":80", handler))
}
