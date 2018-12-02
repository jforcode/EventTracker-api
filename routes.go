package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func HealthCheckHandler(env *Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Alive!")
	}
}

func GetEventsHandler(env *Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		events, err := env.EventsHandler.GetAllEvents()
		if err != nil {
			HandleHttpError(w, err)
		}

		resp := EventsResponse{events}
		eventJson, err := json.Marshal(resp)
		if err != nil {
			HandleHttpError(w, err)
		}

		io.WriteString(w, string(eventJson))
	}
}

func GetEventHandler(env *Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		eventId := vars["eventId"]

		event, err := env.EventsHandler.GetEvent(eventId)
		if err != nil {
			HandleHttpError(w, err)
		}

		resp := EventResponse{event}
		eventJson, err := json.Marshal(resp)
		if err != nil {
			HandleHttpError(w, err)
		}

		io.WriteString(w, string(eventJson))
	}
}

func CreateEventHandler(env *Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var event = &Event{}
		post, err := ioutil.ReadAll(r.Body)
		if err != nil {
			HandleHttpError(w, err)
			return
		}

		err = json.Unmarshal(post, event)
		if err != nil {
			HandleHttpError(w, err)
			return
		}

		eventId, err := env.EventsHandler.CreateEvent(event)
		if err != nil {
			HandleHttpError(w, err)
			return
		}

		respJson := EventIdResponse{eventId}
		resp, err := json.Marshal(respJson)
		if err != nil {
			HandleHttpError(w, err)
			return
		}

		io.WriteString(w, string(resp))
	}
}
