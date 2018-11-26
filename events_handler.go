package main

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jforcode/DeepError"
	"github.com/jforcode/Util"
)

const (
	GET_EVENT_QUERY = `
		SELECT _id, id, title, note, timestamp, created_at, updated_at, status
		FROM events`
	CREATE_EVENT_QUERY = `
		INSERT INTO events (id, title, note, timestamp, created_at, udpated_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
)

type IEventsHandler interface {
	GetAllEvents() ([]*Event, error)
	GetEvent(eventId string) (*Event, error)
	CreateEvent(*Event) (string, error)
}

type EventsHandler struct {
	db *sql.DB
}

func (handler *EventsHandler) Init(db *sql.DB) {
	handler.db = db
}

func (handler *EventsHandler) GetAllEvents() ([]*Event, error) {
	fn := "GetEvents"

	rows, err := handler.db.Query(GET_EVENT_QUERY)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	events := make([]*Event, 0)
	if rows.Next() {
		event, err := handler.getEventFromDb(rows)
		if err != nil {
			return nil, deepError.New(fn, "scan", err)
		}

		events = append(events, event)
	}

	return events, nil
}

func (handler *EventsHandler) GetEvent(eventId string) (*Event, error) {
	fn := "GetEvent"

	rows, err := handler.db.Query(GET_EVENT_QUERY)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	if rows.Next() {
		event, err := handler.getEventFromDb(rows)
		if err != nil {
			return nil, deepError.New(fn, "scan", err)
		}

		return event, nil
	}

	return nil, deepError.New(fn, "", errors.New("Record not found"))
}

func (handler *EventsHandler) CreateEvent(evt *Event) (string, error) {
	fn := "CreateEvent"

	evt.Id = "asdf"
	params := []interface{}{evt.Id, evt.Title, evt.Note, evt.Timestamp, evt.Type, evt.Tags}

	_, err := util.Db.PrepareAndExec(handler.db, CREATE_EVENT_QUERY, params...)
	if err != nil {
		return "", deepError.New(fn, "prepare and exec", err)
	}

	return evt.Id, nil
}

func (handler *EventsHandler) getEventFromDb(rows *sql.Rows) (*Event, error) {
	fn := "getEventFromDb"

	event := Event{}
	err := rows.Scan(&event.DbId, &event.Id, &event.Title, &event.Note, &event.Timestamp, &event.Type, &event.Tags, &event.CreatedAt, &event.UpdatedAt, &event.Status)

	if err != nil {
		return nil, deepError.New(fn, "scan", err)
	}

	return &event, nil
}
