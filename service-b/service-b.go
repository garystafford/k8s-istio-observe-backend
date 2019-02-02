package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Trace struct {
	ID          string    `json:"id,omitempty"`
	ServiceName string    `json:"serviceName,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
}

var traces []Trace

func Orchestrator(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(500 * time.Millisecond)

	traces = nil
	CallNextService("http://localhost:8003/ping") // service-d

	tmpTrace := Trace{ID: uuid.New().String(), ServiceName: "Service-B", CreatedAt: time.Now().Local()}

	traces = append(traces, tmpTrace)
	fmt.Println(traces)

	json.NewEncoder(w).Encode(traces)
}

func CallNextService(url string) {
	var tmpTraces []Trace
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		err := json.Unmarshal(data, &tmpTraces)
		if err != nil {
			panic(err)
		}

		for _, r := range tmpTraces {
			traces = append(traces, r)
		}
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/ping", Orchestrator).Methods("GET")
	log.Fatal(http.ListenAndServe(":8001", router))
}
