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
	httpGET  = "GET"
	httpPOST = "POST"
)

const (
	paramEventID = "{eventID}"
)

const (
	routeGetHealth   = "/health"
	routeGetEvents   = "/events"
	routeGetEvent    = "/event/" + paramEventID
	routeGetEventF   = "/event/%s"
	routeCreateEvent = "/event"
)

// EventIDResponse represents the response to send back to client, in case of create event
type EventIDResponse struct {
	EventID string `json:"eventID"`
}

// EventResponse represents the response to send to client, in case of a get event
type EventResponse struct {
	Event *Event `json:"event"`
}

// EventsResponse represents the response to send to client, in case of a get all events call
type EventsResponse struct {
	Events []*Event `json:"events"`
}

type env struct {
	EventsHandler IEventsHandler
}

func main() {
	p := properties.MustLoadFile("app.properties", properties.UTF8)

	url := p.GetString("url", "")
	db, err := getDbFromProps(p)
	if err != nil {
		panic(err)
	}

	evtHandler := &EventsHandler{}
	evtHandler.Init(db)
	env := &env{
		evtHandler,
	}

	router := mux.NewRouter()

	router.HandleFunc(routeGetHealth, HealthCheckHandler(env)).Methods(httpGET)
	router.HandleFunc(routeGetEvents, GetEventsHandler(env)).Methods(httpGET)
	router.HandleFunc(routeGetEvent, GetEventHandler(env)).Methods(httpGET)
	router.HandleFunc(routeCreateEvent, CreateEventHandler(env)).Methods(httpPOST)

	log.Fatal(http.ListenAndServe(url, router))
}

func getDbFromProps(p *properties.Properties) (*sql.DB, error) {
	user := p.GetString("user", "")
	password := p.GetString("password", "")
	host := p.GetString("host", "")
	database := p.GetString("db", "")
	flags := make(map[string]string)
	flags["parseTime"] = "true"

	return util.Db.GetDb(user, password, host, database, flags)
}

func handleHTTPError(w http.ResponseWriter, err error) {
	fmt.Printf("Http error occured: %s\n", err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
