package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func invokeRequest(server *Server, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.requestHandler)
	handler.ServeHTTP(rr, req)
	return rr
}

func InvokeBasicRequest(server *Server, method string, url string, body string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	req.RemoteAddr = "172.17.0.1:59766"
	if err != nil {
		log.Fatal(err)
	}
	return invokeRequest(server, req)
}

func ExampleInvokeBasicRequest() {
	var s Server
	InvokeBasicRequest(&s, "POST", "/", `{"node": "10.44.131.21", "seq": 1, "version": 1, "time": 1576815603.449015, "data": {}, "event": "eventsink_started"}`)
	InvokeBasicRequest(&s, "POST", "/", `{"node": "10.44.131.22", "seq": 3, "version": 1, "time": 1576815614.906535, "data": {"protocol": "H323", "is_presenting": false, "connect_time": null, "service_tag": "", "conversation_id": "c6a11ca7-99b4-4bf1-95d6-fdaab5ec7c0a", "media_node": "10.44.131.22", "vendor": "PEXIP (Pexip Infinity Conferencing Platform Pexip Infinity Conferencing Platform/23 (52248.0.0))", "conference": "meet.qa", "source_alias": "pexep_67_ep4_Jordan@vp.pexip.com", "display_name": "pexep_67_ep4_Jordan@vp.pexip.com", "uuid": "d5fc4ce0-d0cc-4582-8c67-c033ff397c72", "signalling_node": "10.44.131.22", "call_id": "c6a11ca7-99b4-4bf1-95d6-fdaab5ec7c0a", "role": "chair", "service_type": "connecting", "destination_alias": "meet.qa@10.44.131.22:1720", "is_streaming": false, "is_muted": false, "remote_address": "10.44.100.67", "has_media": true, "system_location": "location 1", "call_direction": "in"}, "event": "participant_connected"}`)
	rr := InvokeBasicRequest(&s, "GET", "/events?clear=1", "")
	fmt.Print(rr.Code, ", ", rr.Body.String())
	rr2 := InvokeBasicRequest(&s, "GET", "/events", "")
	fmt.Print(rr2.Code, ", ", rr2.Body.String())

	// Output:
	// 200, [{"host":"172.17.0.1","hostport":"59766","node":"10.44.131.21","seq":1,"version":1,"time":1576815603.449015,"data":{},"event":"eventsink_started"},{"host":"172.17.0.1","hostport":"59766","node":"10.44.131.22","seq":3,"version":1,"time":1576815614.906535,"data":{"call_direction":"in","call_id":"c6a11ca7-99b4-4bf1-95d6-fdaab5ec7c0a","conference":"meet.qa","connect_time":null,"conversation_id":"c6a11ca7-99b4-4bf1-95d6-fdaab5ec7c0a","destination_alias":"meet.qa@10.44.131.22:1720","display_name":"pexep_67_ep4_Jordan@vp.pexip.com","has_media":true,"is_muted":false,"is_presenting":false,"is_streaming":false,"media_node":"10.44.131.22","protocol":"H323","remote_address":"10.44.100.67","role":"chair","service_tag":"","service_type":"connecting","signalling_node":"10.44.131.22","source_alias":"pexep_67_ep4_Jordan@vp.pexip.com","system_location":"location 1","uuid":"d5fc4ce0-d0cc-4582-8c67-c033ff397c72","vendor":"PEXIP (Pexip Infinity Conferencing Platform Pexip Infinity Conferencing Platform/23 (52248.0.0))"},"event":"participant_connected"}]
	// 200, []
}

func TestInvalidGetPath(t *testing.T) {
	var s Server
	rr := InvokeBasicRequest(&s, "GET", "/favicon.ico", "")
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
	expected := "Invalid GET path /favicon.ico\n"
	if actual := rr.Body.String(); actual != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}
}

func TestGetEmpty(t *testing.T) {
	var s Server
	rr := InvokeBasicRequest(&s, "GET", "/events", "")

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := "[]\n"
	if actual := rr.Body.String(); actual != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}

	if value := rr.Header().Get("Content-type"); value != "text/json" {
		t.Errorf("bad content type %v", value)
	}
}
