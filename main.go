package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
)

// Event represents a management event
type Event struct {
	Host     string    `json:"host"`
	HostPort string    `json:"hostport"`
	Node     string    `json:"node"`
	Seq      int64     `json:"seq"`
	Version  int8      `json:"version"`
	Time     float64   `json:"time"`
	Data     EventData `json:"data"`
	Event    string    `json:"event"`
}

// EventData is an arbitrary mapping of key=value
type EventData map[string]interface{}

// EventStore is a thread-safe list of events
type EventStore struct {
	sync.Mutex
	events []Event
}

// AddEvent adds an event to the store
func (es *EventStore) AddEvent(event Event) {
	es.Lock()
	defer es.Unlock()
	es.events = append(es.events, event)
	log.Printf("Now have %d events", len(es.events))
}

// GetEvents returns the events in the store
func (es *EventStore) GetEvents() []Event {
	es.Lock()
	defer es.Unlock()
	ret := make([]Event, len(es.events))
	copy(ret, es.events)
	return ret
}

// GetAndClearEvents clears and returns the events
func (es *EventStore) GetAndClearEvents() []Event {
	es.Lock()
	defer es.Unlock()
	ret := es.events
	es.events = []Event{}
	return ret
}

// Server holds the state of the server
type Server struct {
	es EventStore
}

func encodeEvents(w io.Writer, events []Event) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(events)
}

func decodeEvent(r io.Reader, event *Event) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder.Decode(event)
}

func (server *Server) requestHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Handling %s to url %s from %s", req.Method, req.URL, req.RemoteAddr)
	if req.Method == "GET" {
		if req.URL.Path != "/events" {
			http.Error(w, "Invalid GET path "+req.URL.Path, http.StatusBadRequest)
			return
		}
		var events []Event
		_, clear := req.URL.Query()["clear"]
		if clear {
			events = server.es.GetAndClearEvents()
		} else {
			events = server.es.GetEvents()
		}
		log.Printf("Returning %d events...\n", len(events))
		w.Header().Set("Content-type", "text/json")
		err := encodeEvents(w, events)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if req.Method == "POST" {
		if req.URL.Path != "/" {
			http.Error(w, "Invalid POST path "+req.URL.Path, http.StatusBadRequest)
			return
		}
		defer req.Body.Close()
		// TODO: more input validation on post body data
		var event Event
		err := decodeEvent(req.Body, &event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ip, port, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		event.Host = ip
		event.HostPort = port
		log.Printf("Received event %+v\n", event)
		server.es.AddEvent(event)
	}
}

// StartServer spins up a http server, returns error if it fails
func StartServer(port int, useTLS bool) error {
	log.Printf("Starting event server (port=%d, useTLS=%v)\n", port, useTLS)
	var s Server
	http.HandleFunc("/", s.requestHandler)
	addr := fmt.Sprintf(":%d", port)
	if useTLS {
		return http.ListenAndServeTLS(
			addr,
			"/etc/certs/server.pem",
			"/etc/certs/privatekey.pem",
			nil)
	}
	return http.ListenAndServe(addr, nil)
}

func main() {
	port := flag.Int("port", 8000, "Port to listen on")
	useTLS := flag.Bool("use-tls", false, "Use TLS")
	flag.Parse()
	log.Fatal(StartServer(*port, *useTLS))
}
