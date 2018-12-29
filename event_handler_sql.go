package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jforcode/Go-DeepError"
)

var (
	queryGetEvents = fmt.Sprintf(`
		SELECT E.%s, E.%s, E.%s, E.%s, E.%s, E.%s, E.%s, E.%s, E.%s
		FROM %s E`,
		colDbID, eventsColID, eventsColTitle, eventsColNote, eventsColCreatedAt, eventsColTypeID, colCreatedAt, colUpdatedAt, colStatus,
		eventsTableName)

	queryGetEvent = fmt.Sprintf(`
		SELECT E.%s, E.%s, E.%s, E.%s, E.%s, E.%s, E.%s, E.%s, E.%s
		FROM %s E
		WHERE E.%s = ?`,
		colDbID, eventsColID, eventsColTitle, eventsColNote, eventsColCreatedAt, eventsColTypeID, colCreatedAt, colUpdatedAt, colStatus,
		eventsTableName,
		eventsColID)

	queryGetEventType = fmt.Sprintf(`
		SELECT ET.%s, ET.%s, ET.%s, ET.%s, ET.%s
		FROM %s ET
		WHERE ET.%s IN (%s)`,
		colDbID, eventTagsColValue, colCreatedAt, colUpdatedAt, colStatus,
		eventTypesTableName,
		colDbID, "%s")

	queryGetEventTags = fmt.Sprintf(`
		SELECT ETG.%s, ETG.%s, ETG.%s, ETG.%s, ETG.%s, ETM.%s
		FROM %s ETG
		JOIN %s ETM ON ETM.%s = ETG.%s
		WHERE ETM.%s IN (%s)`,
		colDbID, eventTagsColValue, colCreatedAt, colUpdatedAt, colStatus, eventTagMapColEventID,
		eventTagsTableName,
		eventTagMapTableName, eventTagMapColTagID, colDbID,
		eventTagMapColEventID, "%s")

	queryCreateEvent = fmt.Sprintf(`
		INSERT INTO events (%s, %s, %s, %s, %s)
		VALUES (?, ?, ?, ?, ?)`,
		eventsColID, eventsColTitle, eventsColNote, eventsColCreatedAt, eventsColTypeID)
)

// IEventsHandler is the common interface to use for events business logic
type IEventsHandler interface {
	GetAllEvents() ([]*Event, error)
	GetEvent(eventID string) (*Event, error)
	CreateEvent(*Event) (string, error)
}

// EventsHandler is a concrete event handler for mysql
type EventsHandler struct {
	db      *sql.DB
	dbStuff *dbStuff
}

// Init initialises the handler
func (handler *EventsHandler) Init(db *sql.DB) {
	handler.dbStuff = &dbStuff{db}
	handler.db = db
}

// GetAllEvents gets all events
func (handler *EventsHandler) GetAllEvents() ([]*Event, error) {
	fn := "GetEvents"

	rows, err := handler.db.Query(queryGetEvents)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	events, err := handler.getEventsFromRows(rows)
	if err != nil {
		return nil, deepError.New(fn, "getEventsFromDb", err)
	}

	return events, nil
}

// GetEvent finds an event by event id
func (handler *EventsHandler) GetEvent(eventID string) (*Event, error) {
	fn := "GetEvent"

	rows, err := handler.db.Query(queryGetEvent, eventID)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	events, err := handler.getEventsFromRows(rows)
	if err != nil {
		return nil, deepError.New(fn, "getEventsFromDb", err)
	}

	if len(events) == 0 {
		return nil, nil
	}

	return events[0], nil
}

// CreateEvent creates an event
// TODO: transactions
func (handler *EventsHandler) CreateEvent(event *Event) (string, error) {
	fn := "CreateEvent"

	eventType, err := handler.findOrCreateEventType(event.Type.Value)
	if err != nil {
		return "", deepError.New(fn, "find or create event type", err)
	}
	event.Type = eventType

	updatedEventTags := make([]*EventTag, len(event.Tags))
	for index, eventTag := range event.Tags {
		foundEventTag, err2 := handler.findOrCreateEventTag(eventTag.Value)
		if err2 != nil {
			return "", deepError.New(fn, "find or create event tag", err2)
		}
		updatedEventTags[index] = foundEventTag
	}
	event.Tags = updatedEventTags

	event.ID = uuid.New().String()
	event.UserCreatedAt = event.UserCreatedAt.UTC()
	eventDbID, err3 := handler.dbStuff.insertEvent(event)
	if err3 != nil {
		return "", deepError.New(fn, "insert event", err3)
	}
	event.DbID = eventDbID

	for _, eventTag := range event.Tags {
		eventTagMap := &EventTagMap{EventID: event.DbID, TagID: eventTag.DbID}
		_, err4 := handler.dbStuff.insertEventTagMapping(eventTagMap)
		if err4 != nil {
			return "", deepError.New(fn, "insert event tag map", err4)
		}
	}

	return event.ID, nil
}

func (handler *EventsHandler) findOrCreateEventType(value string) (*EventType, error) {
	fn := "findOrCreateEventType"

	eventType, err := handler.dbStuff.findEventTypeByValue(value)
	if err != nil {
		return nil, deepError.New(fn, "find event type", err)
	}

	if eventType == nil {
		eventType = &EventType{Value: value}
		eventTypeDbID, err2 := handler.dbStuff.insertEventType(eventType)
		if err2 != nil {
			return nil, deepError.New(fn, "insert event type", err)
		}
		eventType.DbID = eventTypeDbID
	}

	return eventType, nil
}

func (handler *EventsHandler) findOrCreateEventTag(value string) (*EventTag, error) {
	fn := "findOrCreateEventTag"

	eventTag, err := handler.dbStuff.findEventTagByValue(value)
	if err != nil {
		return nil, deepError.New(fn, "find event tag by value", err)
	}

	if eventTag == nil {
		eventTag = &EventTag{Value: value}
		eventTagDbID, err := handler.dbStuff.insertEventTag(eventTag)
		if err != nil {
			return nil, deepError.New(fn, "insert event tag", err)
		}
		eventTag.DbID = eventTagDbID
	}

	return eventTag, nil
}

func (handler *EventsHandler) getEventsFromRows(rows *sql.Rows) ([]*Event, error) {
	fn := "getEventsFromDb"

	events := make([]*Event, 0)

	for rows.Next() {
		event := &Event{}
		event.Type = &EventType{}
		event.Tags = make([]*EventTag, 0)

		err := rows.Scan(&event.DbID, &event.ID, &event.Title, &event.Note, &event.UserCreatedAt, &event.Type.DbID, &event.CreatedAt, &event.UpdatedAt, &event.Status)
		if err != nil {
			return nil, deepError.New(fn, "scan", err)
		}

		typeID := event.Type.DbID
		eventType, err := handler.dbStuff.findEventTypeByID(typeID)
		if err != nil {
			return nil, deepError.New(fn, "find event type by id", err)
		}

		event.Type = eventType

		eventTags, err := handler.dbStuff.findEventTagsByEventID(event.ID)
		if err != nil {
			return nil, deepError.New(fn, "find event tags by event id", err)
		}

		event.Tags = eventTags

		events = append(events, event)
	}

	return events, nil
}
