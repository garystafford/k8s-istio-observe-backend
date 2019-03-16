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
	traces = nil
	CallNextService("http://service-b/api/ping", w, r)
	CallNextService("http://service-c/api/ping", w, r)

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

func CallNextService(url string, w http.ResponseWriter, r *http.Request) {
	log.Info(url)
	var tmpTraces []Trace

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
	}

	req.Header.Add("x-request-id", r.Header.Get("x-request-id"))
	req.Header.Add("x-b3-traceid", r.Header.Get("x-b3-traceid"))
	req.Header.Add("x-b3-spanid", r.Header.Get("x-b3-spanid"))
	req.Header.Add("x-b3-parentspanid", r.Header.Get("x-b3-parentspanid"))
	req.Header.Add("x-b3-sampled", r.Header.Get("x-b3-sampled"))
	req.Header.Add("x-b3-flags", r.Header.Get("x-b3-flags"))
	req.Header.Add("x-ot-span-context", r.Header.Get("x-ot-span-context"))
	log.Info(req)

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		log.Error(err)
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(body, &tmpTraces)
	if err != nil {
		log.Error(err)
	}

	for _, r := range tmpTraces {
		traces = append(traces, r)
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
