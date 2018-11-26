package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jforcode/Util"

	"github.com/gorilla/mux"
	"github.com/magiconair/properties"
)

const (
	HTTP_GET  = "GET"
	HTTP_POST = "POST"
)

const (
	PARAM_EVENT_ID = "{eventId}"
)

const (
	ROUTE_GET_HEALTH   = "/health"
	ROUTE_GET_EVENTS   = "/events"
	ROUTE_GET_EVENT    = "/event/" + PARAM_EVENT_ID
	ROUTE_GET_EVENT_F  = "/event/%s"
	ROUTE_CREATE_EVENT = "/event"
)

type EventIdResponse struct {
	EventId string `json:"eventId"`
}

type EventResponse struct {
	Event *Event `json:"event"`
}

type EventsResponse struct {
	Events []*Event `json:"events"`
}

type Env struct {
	EventsHandler IEventsHandler
}

func main() {
	p := properties.MustLoadFile("app.properties", properties.UTF8)

	url := p.GetString("url", "")

	user := p.GetString("user", "")
	password := p.GetString("password", "")
	host := p.GetString("host", "")
	database := p.GetString("db", "")
	flags := make(map[string]string)
	flags["parseTime"] = "true"

	db, err := util.Db.GetDb(user, password, host, database, flags)
	if err != nil {
		panic(err)
	}

	evtHandler := &EventsHandler{}
	evtHandler.Init(db)
	env := Env{
		evtHandler,
	}

	router := mux.NewRouter()

	router.HandleFunc(ROUTE_GET_HEALTH, HealthCheckHandler(env)).Methods(HTTP_GET)
	router.HandleFunc(ROUTE_GET_EVENTS, GetEventsHandler(env)).Methods(HTTP_GET)
	router.HandleFunc(ROUTE_GET_EVENT, GetEventHandler(env)).Methods(HTTP_GET)
	router.HandleFunc(ROUTE_CREATE_EVENT, CreateEventHandler(env)).Methods(HTTP_POST)

	log.Fatal(http.ListenAndServe(url, router))
}

func HealthCheckHandler(env Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Alive!")
	}
}

func GetEventsHandler(env Env) func(w http.ResponseWriter, r *http.Request) {
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

func GetEventHandler(env Env) func(w http.ResponseWriter, r *http.Request) {
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

func CreateEventHandler(env Env) func(w http.ResponseWriter, r *http.Request) {
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

func HandleHttpError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
