package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

// HealthCheckHandler is an api route to just check the health of the api.
func HealthCheckHandler(env *env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Alive!")
	}
}

// GetEventsHandler is a route to return all the events available.
// NOTE: unpaginated. unauthenticated
func GetEventsHandler(env *env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		events, err := env.EventsHandler.GetAllEvents()
		if err != nil {
			handleHTTPError(w, err)
		}

		resp := EventsResponse{events}
		eventJSON, err := json.Marshal(resp)
		if err != nil {
			handleHTTPError(w, err)
		}

		io.WriteString(w, string(eventJSON))
	}
}

// GetEventHandler is a route to return a specific event based on the event id
// TODO: if no event found, return error
func GetEventHandler(env *env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		eventID := vars["eventID"]

		event, err := env.EventsHandler.GetEvent(eventID)
		if err != nil {
			handleHTTPError(w, err)
		}

		resp := EventResponse{event}
		eventJSON, err := json.Marshal(resp)
		if err != nil {
			handleHTTPError(w, err)
		}

		io.WriteString(w, string(eventJSON))
	}
}

// CreateEventHandler is a route to create a new event in the system.
// Not an update call, will decide later if to create new, or use this only for update
func CreateEventHandler(env *env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var event = &Event{}
		post, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleHTTPError(w, err)
			return
		}

		err = json.Unmarshal(post, event)
		if err != nil {
			handleHTTPError(w, err)
			return
		}

		eventID, err := env.EventsHandler.CreateEvent(event)
		if err != nil {
			handleHTTPError(w, err)
			return
		}

		respJSON := EventIDResponse{eventID}
		resp, err := json.Marshal(respJSON)
		if err != nil {
			handleHTTPError(w, err)
			return
		}

		io.WriteString(w, string(resp))
	}
}
