package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
)

func TestHealthCheck(t *testing.T) {
	req, err := http.NewRequest(HTTP_GET, ROUTE_GET_HEALTH, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	env := Env{}

	router.HandleFunc(ROUTE_GET_HEALTH, HealthCheckHandler(env))
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
	router := mux.NewRouter()
	env := Env{
		&TestEventHandler{},
	}

	router.HandleFunc(ROUTE_CREATE_EVENT, CreateEventHandler(env)).Methods(HTTP_POST)
	router.HandleFunc(ROUTE_GET_EVENT, GetEventHandler(env)).Methods(HTTP_GET)

	eventJson := `
		{
			"title": "Test Event",
			"note": "Some Test note",
			"timestamp": "2018-11-25T11:26:08+00:00",
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

	eventId, err := CreateEvent(router, eventJson)
	if err != nil {
		t.Fatalf("Error while creating event: %+v", err.Error())
	}

	actual, err := GetEvent(router, eventId)
	if err != nil {
		t.Fatalf("Error while getting event: %+v", err.Error())
	}

	expectedTime, err := time.Parse(time.RFC3339, "2018-11-25T11:26:08+00:00")
	if err != nil {
		t.Fatalf("Failed to parse time: %+v", "2018-11-25T11:26:08+00:00")
	}

	expected := &Event{
		Id:        eventId,
		Title:     "Test Event",
		Note:      "Some Test note",
		Timestamp: expectedTime,
		Type:      &EventType{Value: "start"},
		Tags:      []*EventTag{{Value: "test1"}, {Value: "test2"}},
	}

	if !cmp.Equal(expected, actual) {
		t.Errorf("Didnt' get event as expected\nExpected: %+v\nGot: %+v", expected, actual)
	}
}

func CreateEvent(router *mux.Router, eventJson string) (string, error) {
	req, err := http.NewRequest(HTTP_POST, ROUTE_CREATE_EVENT, strings.NewReader(eventJson))
	if err != nil {
		return "", err
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		return "", errors.New(fmt.Sprintf("Status Not OK. Got status code: %d", status))
	}

	resp, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		return "", err
	}

	eventIdResp := EventIdResponse{}
	err = json.Unmarshal(resp, &eventIdResp)
	if err != nil {
		return "", err
	}

	return eventIdResp.EventId, nil
}

func GetEvent(router *mux.Router, eventId string) (*Event, error) {
	path := fmt.Sprintf(ROUTE_GET_EVENT_F, eventId)
	req, err := http.NewRequest(HTTP_GET, path, nil)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Status Not OK. Got status code: %d", status))
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
	req, err := http.NewRequest(HTTP_GET, ROUTE_GET_EVENTS, nil)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Status Not OK. Got status code: %d", status))
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

type TestEventHandler struct {
	events      []*Event
	lastEventId int
}

func (handler *TestEventHandler) GetAllEvents() ([]*Event, error) {
	return handler.events, nil
}

func (handler *TestEventHandler) GetEvent(eventId string) (*Event, error) {
	for _, evt := range handler.events {
		if strings.EqualFold(eventId, evt.Id) {
			return evt, nil
		}
	}

	return nil, errors.New("Event with id " + eventId + " not found")
}

func (handler *TestEventHandler) CreateEvent(evt *Event) (string, error) {
	handler.lastEventId++
	evt.Id = strconv.Itoa(handler.lastEventId)
	handler.events = append(handler.events, evt)
	return evt.Id, nil
}
