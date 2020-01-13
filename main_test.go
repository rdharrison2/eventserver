package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const event = `{"node": "10.44.131.21", "seq": 1, "version": 1, "time": 1576815603.449015, "data": {}, "event": "eventsink_started"}`
const unknownField = `{"node": "10.44.131.22", "sally": 14, "version": 1, "time": 1576815642.496487, "data": {"protocol": "H323"}, "event": "participant_disconnected"}`
const unicodeEvent = `{"node": "10.44.131.22", "seq": 14, "version": 1, "time": 1576815642.496487, "data": {"protocol": "H323", "destination_alias": "sip:日本語@10.44.131.21"}, "event": "participant_disconnected"}`

func InvokeRequest(server *Server, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h := server.handler()
	h.ServeHTTP(rr, req)
	return rr
}

func request(method, url, body string) *http.Request {
	bodyReader := strings.NewReader(body)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		log.Fatal(err)
	}
	req.RemoteAddr = "172.17.0.1:59766"
	return req
}

func ExampleInvokeRequest() {
	var s Server
	InvokeRequest(&s, request("POST", "/", `{"node": "10.44.131.21", "seq": 1, "version": 1, "time": 1576815603.449015, "data": {}, "event": "eventsink_started"}`))
	InvokeRequest(&s, request("POST", "/", `{"node": "10.44.131.22", "seq": 3, "version": 1, "time": 1576815614.906535, "data": {"protocol": "H323", "is_presenting": false, "connect_time": null, "service_tag": "", "conversation_id": "c6a11ca7-99b4-4bf1-95d6-fdaab5ec7c0a", "media_node": "10.44.131.22", "vendor": "PEXIP (Pexip Infinity Conferencing Platform Pexip Infinity Conferencing Platform/23 (52248.0.0))", "conference": "meet.qa", "source_alias": "pexep_67_ep4_Jordan@vp.pexip.com", "display_name": "pexep_67_ep4_Jordan@vp.pexip.com", "uuid": "d5fc4ce0-d0cc-4582-8c67-c033ff397c72", "signalling_node": "10.44.131.22", "call_id": "c6a11ca7-99b4-4bf1-95d6-fdaab5ec7c0a", "role": "chair", "service_type": "connecting", "destination_alias": "meet.qa@10.44.131.22:1720", "is_streaming": false, "is_muted": false, "remote_address": "10.44.100.67", "has_media": true, "system_location": "location 1", "call_direction": "in"}, "event": "participant_connected"}`))
	rr := InvokeRequest(&s, request("GET", "/events?clear=1", ""))
	fmt.Print(rr.Code, ", ", rr.Body.String())
	rr2 := InvokeRequest(&s, request("GET", "/events", ""))
	fmt.Print(rr2.Code, ", ", rr2.Body.String())

	// Output:
	// 200, [{"host":"172.17.0.1","hostport":"59766","node":"10.44.131.21","seq":1,"version":1,"time":1576815603.449015,"data":{},"event":"eventsink_started"},{"host":"172.17.0.1","hostport":"59766","node":"10.44.131.22","seq":3,"version":1,"time":1576815614.906535,"data":{"call_direction":"in","call_id":"c6a11ca7-99b4-4bf1-95d6-fdaab5ec7c0a","conference":"meet.qa","connect_time":null,"conversation_id":"c6a11ca7-99b4-4bf1-95d6-fdaab5ec7c0a","destination_alias":"meet.qa@10.44.131.22:1720","display_name":"pexep_67_ep4_Jordan@vp.pexip.com","has_media":true,"is_muted":false,"is_presenting":false,"is_streaming":false,"media_node":"10.44.131.22","protocol":"H323","remote_address":"10.44.100.67","role":"chair","service_tag":"","service_type":"connecting","signalling_node":"10.44.131.22","source_alias":"pexep_67_ep4_Jordan@vp.pexip.com","system_location":"location 1","uuid":"d5fc4ce0-d0cc-4582-8c67-c033ff397c72","vendor":"PEXIP (Pexip Infinity Conferencing Platform Pexip Infinity Conferencing Platform/23 (52248.0.0))"},"event":"participant_connected"}]
	// 200, []
}

func TestGetRequests(t *testing.T) {
	var tests = []struct {
		testname      string
		authenticated bool
		path          string
		code          int
		response      string
	}{
		{"events", false, "/events", http.StatusOK, "[]\n"},
		{"events_authenticated", true, "/events", http.StatusOK, "[]\n"},
		{"bad path", false, "/favicon.ico", http.StatusBadRequest, "Invalid GET path /favicon.ico\n"},
		{"bad path authenticated", true, "/favicon.ico", http.StatusUnauthorized, "You are Unauthorized to access the application.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			var s Server
			if tt.authenticated {
				s.userpass = &Credential{Username: "admin", Password: "PEXNOTE"}
			}
			req := request("GET", tt.path, "")
			rr := InvokeRequest(&s, req)
			if status := rr.Code; status != tt.code {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.code)
			}
			if response := rr.Body.String(); response != tt.response {
				t.Errorf("handler returned unexpected body: got '%v' want '%v'",
					response, tt.response)
			}
			if rr.Code == http.StatusOK {
				if value := rr.Header().Get("Content-type"); value != "text/json" {
					t.Errorf("bad content type %v", value)
				}
			}
		})
	}
}

func TestPostRequests(t *testing.T) {
	var tests = []struct {
		path, data string
		code       int
		response   string
	}{
		{"/", event, http.StatusOK, ""},
		{"/", unicodeEvent, http.StatusOK, ""},
		{"/events", event, http.StatusBadRequest, "Invalid POST path /events\n"},
		{"/", "", http.StatusBadRequest, "EOF\n"},
		{"/", "footle", http.StatusBadRequest, "invalid character 'o' in literal false (expecting 'a')\n"},
		{"/", unknownField, http.StatusBadRequest, "json: unknown field \"sally\"\n"},
	}
	for i, tt := range tests {
		testname := fmt.Sprintf("%d:%v", i, tt.path)
		t.Run(testname, func(t *testing.T) {
			var s Server
			req := request("POST", tt.path, tt.data)
			rr := InvokeRequest(&s, req)
			if status := rr.Code; status != tt.code {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.code)
			}
			if response := rr.Body.String(); response != tt.response {
				t.Errorf("handler returned unexpected body: got '%v' want '%v'",
					response, tt.response)
			}
			expectedEvents := 0
			if rr.Code == http.StatusOK {
				expectedEvents = 1
			}
			if numEvents := len(s.es.events); numEvents != expectedEvents {
				t.Errorf("Found %d events but expected %d", numEvents, expectedEvents)
			}
		})
	}
}

func TestPostRequestsAuthenticated(t *testing.T) {
	var tests = []struct {
		path, data, username, password string
		code                           int
		response                       string
	}{
		{"/", event, "", "", http.StatusUnauthorized, "You are Unauthorized to access the application.\n"},
		{"/", event, "admin", "footle", http.StatusUnauthorized, "You are Unauthorized to access the application.\n"},
		{"/", event, "admin", "PEXNOTE", http.StatusOK, ""},
	}
	for i, tt := range tests {
		testname := fmt.Sprintf("%d:%v", i, tt.path)
		t.Run(testname, func(t *testing.T) {
			var s Server
			s.userpass = &Credential{Username: "admin", Password: "PEXNOTE"}
			req := request("POST", tt.path, tt.data)
			if tt.username != "" || tt.password != "" {
				req.SetBasicAuth(tt.username, tt.password)
			}
			rr := InvokeRequest(&s, req)
			if status := rr.Code; status != tt.code {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.code)
			}
			if response := rr.Body.String(); response != tt.response {
				t.Errorf("handler returned unexpected body: got '%v' want '%v'",
					response, tt.response)
			}
			expectedEvents := 0
			if rr.Code == http.StatusOK {
				expectedEvents = 1
			}
			if numEvents := len(s.es.events); numEvents != expectedEvents {
				t.Errorf("Found %d events but expected %d", numEvents, expectedEvents)
			}
		})
	}
}

// TestGetAndClearEvents tests concurrent Add/GetAndClear
func TestGetAndClearEvents(t *testing.T) {
	var es EventStore
	// buffered channel to signal adding of events so we know when we've
	// finished adding events
	done := make(chan int, 3)
	go func() {
		for i := 0; i < 100; i++ {
			es.AddEvent(Event{Seq: int64(i)})
			done <- i + 1
		}
	}()
	var events []Event
	var lastDone int
	for lastDone < 100 {
		lastDone = <-done
		newEvents := es.GetAndClearEvents()
		for _, ev := range newEvents {
			events = append(events, ev)
		}
	}
	if len(events) != 100 {
		t.Errorf("Expected 100 events but found %d", len(events))
	}
	for i, ev := range events {
		if int64(i) != ev.Seq {
			t.Errorf("Unexpected Seq for %d th event %v", i, ev)
		}
	}
}
