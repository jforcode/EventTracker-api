package main

import (
	"database/sql"
	"fmt"
	"strings"

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
		return nil, deepError.New(fn, "Not found", err)
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
	event.Timestamp = event.Timestamp.UTC()
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

	mapEvents := make(map[int64]Event, 0)
	typeIDs := make([]int64, 0)
	eventIDs := make([]int64, 0)

	for rows.Next() {
		event := Event{}
		event.Type = &EventType{}
		event.Tags = make([]*EventTag, 0)

		err := rows.Scan(&event.DbID, &event.ID, &event.Title, &event.Note, &event.Timestamp, &event.Type.DbID, &event.CreatedAt, &event.UpdatedAt, &event.Status)
		if err != nil {
			return nil, deepError.New(fn, "scan", err)
		}

		mapEvents[event.DbID] = event
		typeIDs = append(typeIDs, event.Type.DbID)
		eventIDs = append(eventIDs, event.DbID)
	}

	eventTypes, err := handler.getTypeIDTypeMappingsFromDb(typeIDs)
	if err != nil {
		return nil, deepError.New(fn, "get type id type mappings", err)
	}

	for _, event := range mapEvents {
		event.Type = eventTypes[event.Type.DbID]
	}

	eventTags, err := handler.getEventIDTagMappingsFromDb(eventIDs)
	if err != nil {
		return nil, deepError.New(fn, "get event id tag mappings", err)
	}

	for eventID, tag := range eventTags {
		event := mapEvents[eventID]
		event.Tags = append(event.Tags, tag)
	}

	events := make([]*Event, len(mapEvents))
	i := 0
	for _, event := range mapEvents {
		events[i] = &event
		i++
	}

	return events, nil
}

func (handler *EventsHandler) getTypeIDTypeMappingsFromDb(typeIDs []int64) (map[int64]*EventType, error) {
	// returns a map of Type id and Type
	fn := "getTypeIDTypeMappingsFromDb"
	lenIDs := len(typeIDs)

	if lenIDs == 0 {
		return map[int64]*EventType{}, nil
	}

	paramsS := ""
	if lenIDs == 1 {
		paramsS = "?"
	} else {
		paramsS = "?" + strings.Repeat(", ?", lenIDs-1)
	}

	query := fmt.Sprintf(queryGetEventType, paramsS)

	params := make([]interface{}, lenIDs)
	for i, typeID := range typeIDs {
		params[i] = typeID
	}

	rows, err := handler.db.Query(query, params...)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	mapTypeIDType := make(map[int64]*EventType, 0)
	for rows.Next() {
		eventType := &EventType{}
		rows.Scan(&eventType.DbID, &eventType.Value, &eventType.CreatedAt, &eventType.UpdatedAt, &eventType.Status)

		mapTypeIDType[eventType.DbID] = eventType
	}

	return mapTypeIDType, nil
}

func (handler *EventsHandler) getEventIDTagMappingsFromDb(eventIDs []int64) (map[int64]*EventTag, error) {
	// returns a map of EventID and Event Tag
	fn := "getEventIDTagMappingsFromDb"
	lenIDs := len(eventIDs)

	if lenIDs == 0 {
		return map[int64]*EventTag{}, nil
	}

	paramsS := ""
	if lenIDs == 1 {
		paramsS = "?"
	} else {
		paramsS = "?" + strings.Repeat(", ?", lenIDs-1)
	}

	query := fmt.Sprintf(queryGetEventTags, paramsS)

	params := make([]interface{}, lenIDs)
	for i, eventID := range eventIDs {
		params[i] = eventID
	}

	rows, err := handler.db.Query(query, params...)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	mapEventIDTag := make(map[int64]*EventTag, 0)
	for rows.Next() {
		eventTag := &EventTag{}
		var eventID int64
		rows.Scan(&eventTag.DbID, &eventTag.Value, &eventTag.CreatedAt, &eventTag.UpdatedAt, &eventTag.Status, &eventID)

		mapEventIDTag[eventID] = eventTag
	}

	return mapEventIDTag, nil
}
