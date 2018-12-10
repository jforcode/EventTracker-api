package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/jforcode/Go-Util"

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
	db, err := GetDbFromProperties(p)
	if err != nil {
		panic(err)
	}

	evtHandler := &EventsHandler{}
	evtHandler.Init(db)
	env := &Env{
		evtHandler,
	}

	router := mux.NewRouter()

	router.HandleFunc(ROUTE_GET_HEALTH, HealthCheckHandler(env)).Methods(HTTP_GET)
	router.HandleFunc(ROUTE_GET_EVENTS, GetEventsHandler(env)).Methods(HTTP_GET)
	router.HandleFunc(ROUTE_GET_EVENT, GetEventHandler(env)).Methods(HTTP_GET)
	router.HandleFunc(ROUTE_CREATE_EVENT, CreateEventHandler(env)).Methods(HTTP_POST)

	log.Fatal(http.ListenAndServe(url, router))
}

func GetDbFromProperties(p *properties.Properties) (*sql.DB, error) {
	user := p.GetString("user", "")
	password := p.GetString("password", "")
	host := p.GetString("host", "")
	database := p.GetString("db", "")
	flags := make(map[string]string)
	flags["parseTime"] = "true"

	return util.Db.GetDb(user, password, host, database, flags)
}

func HandleHttpError(w http.ResponseWriter, err error) {
	fmt.Printf("Http error occured: %s\n", err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
