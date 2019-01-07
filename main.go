package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jforcode/Go-Util"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/magiconair/properties"
)

const (
	paramEventID = "{eventID}"
)

const (
	routeGetHealth   = "/health"
	routeGetEvents   = "/events"
	routeGetEvent    = "/events/" + paramEventID
	routeGetEventF   = "/events/%s"
	routeCreateEvent = "/event"
)

// ResponseError is the error format in case of any error, be it internal or user-defined
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err *ResponseError) Error() string {
	return strconv.Itoa(err.Code) + " : " + err.Message
}

// Response is the final response sent to the client
type Response struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data"`
	Error   *ResponseError `json:"error"`
}

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
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)

	router.HandleFunc(routeGetHealth, HealthCheckHandler(env)).Methods(http.MethodGet)
	router.HandleFunc(routeGetEvents, GetEventsHandler(env)).Methods(http.MethodGet)
	router.HandleFunc(routeGetEvent, GetEventHandler(env)).Methods(http.MethodGet)
	router.HandleFunc(routeCreateEvent, CreateEventHandler(env)).Methods(http.MethodPost)

	log.Fatal(http.ListenAndServe(url, loggedRouter))
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

func handleHTTPSuccess(w http.ResponseWriter, data interface{}) {
	resp := Response{
		Success: true,
		Data:    data,
		Error:   nil,
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		handleHTTPError(w, err)
		return
	}

	io.WriteString(w, string(respJSON))
}

func handleHTTPError(w http.ResponseWriter, err error) {
	fmt.Printf("Http error occured: %s\n", err.Error())
	resp := Response{
		Success: false,
		Data:    nil,
		Error: &ResponseError{
			Code:    0,
			Message: err.Error(),
		},
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		io.WriteString(w, err.Error())
	} else {
		io.WriteString(w, string(respJSON))
	}
}
