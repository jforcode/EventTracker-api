package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
)

func HandleTestError(t *testing.T, from string, err error) {
	if err != nil {
		t.Fatalf("Error in %s\n%s", from, err)
	}
}

func TestHealthCheck(t *testing.T) {
	fn := "TestHealthCheck"

	req, err := http.NewRequest(httpGET, routeGetHealth, nil)
	HandleTestError(t, fn, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	env := &env{}

	router.HandleFunc(routeGetHealth, HealthCheckHandler(env))
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code. Wanted %+v. Got %+v", http.StatusOK, status)
	}

	expected := "Alive!"
	if actual := rr.Body.String(); actual != expected {
		t.Errorf("Handler returned wrong response. Wanted %+v. Got %+v", expected, actual)
	}
}

func TestCreateEvent(t *testing.T) {
	fn := "TestCreateEvent"

	router := mux.NewRouter()
	env := &env{
		&TestEventHandler{},
	}

	router.HandleFunc(routeCreateEvent, CreateEventHandler(env)).Methods(httpPOST)
	router.HandleFunc(routeGetEvent, GetEventHandler(env)).Methods(httpGET)

	eventJSON := `
		{
			"title": "Test Event",
			"note": "Some Test note",
			"created_at": "2018-11-25T11:26:08Z",
			"type": {
				"value": "start"
			},
			"tags": [
				{
					"value": "test1"
				},
				{
					"value": "test2"
				}
			]
		}`

	eventID, err := CreateEvent(router, eventJSON)
	HandleTestError(t, fn, err)

	actual, err := GetEvent(router, eventID)
	HandleTestError(t, fn, err)

	expectedTime, err := time.Parse(time.RFC3339, "2018-11-25T11:26:08Z")
	HandleTestError(t, fn, err)

	expected := &Event{
		ID:            eventID,
		Title:         "Test Event",
		Note:          "Some Test note",
		UserCreatedAt: expectedTime,
		Type:          &EventType{Value: "start"},
		Tags:          []*EventTag{{Value: "test1"}, {Value: "test2"}},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatalf("Didnt' get event as expected\nExpected: %+v\nActual: %+v", expected, actual)
	}
}

func CreateEvent(router *mux.Router, eventJSON string) (string, error) {
	req, err := http.NewRequest(httpPOST, routeCreateEvent, strings.NewReader(eventJSON))
	if err != nil {
		return "", err
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		return "", fmt.Errorf("Status Not OK. Got status code: %d", status)
	}

	resp, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		return "", err
	}

	eventIDResp := EventIDResponse{}
	err = json.Unmarshal(resp, &eventIDResp)
	if err != nil {
		return "", err
	}

	return eventIDResp.EventID, nil
}

func GetEvent(router *mux.Router, eventID string) (*Event, error) {
	path := fmt.Sprintf(routeGetEventF, eventID)
	req, err := http.NewRequest(httpGET, path, nil)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		return nil, fmt.Errorf("Status Not OK. Got status code: %d", status)
	}

	respGet, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		return nil, err
	}

	eventResp := EventResponse{}
	err = json.Unmarshal(respGet, &eventResp)
	if err != nil {
		return nil, err
	}

	return eventResp.Event, nil
}

func GetAllEvents(router *mux.Router) ([]*Event, error) {
	req, err := http.NewRequest(httpGET, routeGetEvents, nil)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		return nil, fmt.Errorf("Status Not OK. Got status code: %d", status)
	}

	respGet, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		return nil, err
	}

	event := []*Event{}
	err = json.Unmarshal(respGet, &event)
	if err != nil {
		return nil, err
	}

	return event, nil
}
