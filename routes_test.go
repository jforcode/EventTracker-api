package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jforcode/Go-Util"
)

func TestHealthCheck(t *testing.T) {
	fn := "TestHealthCheck"

	router := mux.NewRouter()
	router.HandleFunc(routeGetHealth, HealthCheckHandler(nil))

	req, err := http.NewRequest(http.MethodGet, routeGetHealth, nil)
	util.Test.HandleIfTestError(t, err, fn)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	util.Test.AssertEquals(t, http.StatusOK, rr.Code, fn+": Wrong Status Code")

	expected := `
		{
			"success": true,
			"data": "Alive!",
			"error": null
		}`

	util.Test.AssertJSONEquals(t, expected, rr.Body.String(), "Health Check failed")
}

func TestCreateEvent(t *testing.T) {
	fn := "TestCreateEvent"

	router := mux.NewRouter()
	env := &env{
		&TestEventHandler{},
	}

	router.HandleFunc(routeCreateEvent, CreateEventHandler(env)).Methods(http.MethodPost)

	eventJSON := GetTestEventJSON("")

	req, err := http.NewRequest(http.MethodPost, routeCreateEvent, strings.NewReader(eventJSON))
	util.Test.HandleIfTestError(t, err, fn)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	util.Test.AssertEquals(t, http.StatusOK, rr.Code, fn+": Wrong Status Code")
	util.Test.AssertEquals(t, true, strings.Contains(rr.Body.String(), `"success":true`), fmt.Sprintf("Unsuccessful request: %+v", rr.Body))
}

func TestGetEvent(t *testing.T) {
	fn := "TestGetEvent"

	router := mux.NewRouter()
	env := &env{
		&TestEventHandler{},
	}

	router.HandleFunc(routeGetEvent, GetEventHandler(env)).Methods(http.MethodGet)

	eventID, err := env.EventsHandler.CreateEvent(GetTestEvent())
	util.Test.HandleIfTestError(t, err, fn)

	url := fmt.Sprintf(routeGetEventF, eventID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	util.Test.HandleIfTestError(t, err, fn)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	util.Test.AssertEquals(t, http.StatusOK, rr.Code, fn+": Wrong Status Code")

	expected := `
		{
			"success": true,
			"data": {
				"event": ` + GetTestEventJSON(eventID) + `
			},
			"error": null
		}`

	util.Test.AssertJSONEquals(t, expected, rr.Body.String(), "Invalid Event")
}

func TestGetAllEvents(t *testing.T) {
	fn := "TestGetAllEvents"

	router := mux.NewRouter()
	env := &env{
		&TestEventHandler{},
	}

	router.HandleFunc(routeGetEvents, GetEventsHandler(env)).Methods(http.MethodGet)

	eventID1, err := env.EventsHandler.CreateEvent(GetTestEvent())
	util.Test.HandleIfTestError(t, err, fn)
	eventID2, err := env.EventsHandler.CreateEvent(GetTestEvent())
	util.Test.HandleIfTestError(t, err, fn)

	req, err := http.NewRequest(http.MethodGet, routeGetEvents, nil)
	util.Test.HandleIfTestError(t, err, fn)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	util.Test.AssertEquals(t, http.StatusOK, rr.Code, fn+": Wrong Status Code")

	expected := `
		{
			"success": true,
			"data": {
				"events": [
					` + GetTestEventJSON(eventID1) + `,
					` + GetTestEventJSON(eventID2) + `
				]
			},
			"error": null
		}`

	util.Test.AssertJSONEquals(t, expected, rr.Body.String(), "Invalid Event")
}

// thought of refactoring GetTestEvent and GetTestEventJSON and putting as test data
// will do later if required, right now not that much data to make it feasible

func GetTestEvent() *Event {
	mTime, _ := time.Parse(time.RFC3339, "2018-11-25T11:26:08Z")

	return &Event{
		Title: "Test Event",
		Note:  "Some Test note",
		Tags: []*EventTag{
			&EventTag{Value: "test1"},
			&EventTag{Value: "test2"},
		},
		Type:          &EventType{Value: "start"},
		UserCreatedAt: mTime,
	}
}

func GetTestEventJSON(eventID string) string {
	eventIDPart := ""
	if eventID != "" {
		eventIDPart = fmt.Sprintf(`"id": "%s",`, eventID)
	}
	return `{
				` + eventIDPart + `
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
}
