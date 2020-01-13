package main

import (
	"crypto/subtle"
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
	es       EventStore
	userpass *Credential
}

// Credential represents a username/password pair
type Credential struct {
	Username, Password string
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

func BasicAuth(handler http.HandlerFunc, realm string, cred Credential) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		user, pass, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(user), []byte(cred.Username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(cred.Password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("You are Unauthorized to access the application.\n"))
			return
		}
		handler(w, req)
	}
}

// handler returns a new http.ServeMux setup to route requests to server
func (server *Server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", server.requestHandler)
	if server.userpass != nil {
		mux.HandleFunc("/", BasicAuth(server.requestHandler, "events", *server.userpass))
	} else {
		mux.HandleFunc("/", server.requestHandler)
	}
	return mux
}

// StartServer spins up a http server, returns error if it fails
func StartServer(port int, useTLS bool, userpass *Credential) error {
	log.Printf("Starting event server (port=%d, useTLS=%v, userpass=%v)\n", port, useTLS, userpass)
	var s Server
	s.userpass = userpass
	h := s.handler()
	addr := fmt.Sprintf(":%d", port)
	if useTLS {
		return http.ListenAndServeTLS(
			addr,
			"/etc/certs/server.pem",
			"/etc/certs/privatekey.pem",
			h)
	}
	return http.ListenAndServe(addr, h)
}

func main() {
	port := flag.Int("port", 8000, "Port to listen on")
	useTLS := flag.Bool("use-tls", false, "Use TLS")
	username := flag.String("user", "", "Username for auth")
	password := flag.String("password", "", "Password for auth")
	flag.Parse()
	var cred *Credential
	if *username != "" && *password != "" {
		cred = &Credential{Username: *username, Password: *password}
	}
	log.Fatal(StartServer(*port, *useTLS, cred))
}
