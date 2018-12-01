package main

import (
	"database/sql"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jforcode/DeepError"
	"github.com/jforcode/Util"
)

const (
	GET_EVENTS_QUERY = `
		SELECT E._id, E.id, E.title, E.note, E.timestamp, E.type_id, E.created_at, E.updated_at, E.status
		FROM events E`
	GET_EVENT_QUERY = `
		SELECT E._id, E.id, E.title, E.note, E.timestamp, E.type_id, E.created_at, E.updated_at, E.status
		FROM events E
		WHERE E.id = ?`
	GET_EVENT_TYPE_QUERY = `
		SELECT ET._id, ET.value, ET.created_at, ET.updated_at, ET.status
		FROM event_types ET
		WHERE ET._id IN (<params>)`
	GET_EVENT_TAGS_QUERY = `
		SELECT ETG._id, ETG.value, ETG.created_at, ETG.updated_at, ETG.status, ETM.event_id
		FROM event_tags ETG
		JOIN event_tag_mappings ETM ON ETM.tag_id = ETG._id
		WHERE ETM.event_id IN (<params>)`
	CREATE_EVENT_QUERY = `
		INSERT INTO events (id, title, note, timestamp, type, tags)
		VALUES (?, ?, ?, ?, ?, ?)`
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

	rows, err := handler.db.Query(GET_EVENTS_QUERY)
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

func (handler *EventsHandler) GetEvent(eventId string) (*Event, error) {
	fn := "GetEvent"

	rows, err := handler.db.Query(GET_EVENT_QUERY, eventId)
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

func (handler *EventsHandler) CreateEvent(evt *Event) (string, error) {
	fn := "CreateEvent"

	evt.Id = uuid.New().String()
	params := []interface{}{evt.Id, evt.Title, evt.Note, evt.Timestamp, evt.Type, evt.Tags}

	_, err := util.Db.PrepareAndExec(handler.db, CREATE_EVENT_QUERY, params...)
	if err != nil {
		return "", deepError.New(fn, "prepare and exec", err)
	}

	return evt.Id, nil
}

func (handler *EventsHandler) getEventsFromRows(rows *sql.Rows) ([]*Event, error) {
	fn := "getEventsFromDb"

	mapEvents := make(map[int]Event, 0)
	typeIds := make([]int, 0)
	eventIds := make([]int, 0)

	for rows.Next() {
		event := Event{}
		event.Type = &EventType{}
		event.Tags = make([]*EventTag, 0)

		err := rows.Scan(&event.DbId, &event.Id, &event.Title, &event.Note, &event.Timestamp, &event.Type.DbId, &event.CreatedAt, &event.UpdatedAt, &event.Status)
		if err != nil {
			return nil, deepError.New(fn, "scan", err)
		}

		mapEvents[event.DbId] = event
		typeIds = append(typeIds, event.Type.DbId)
		eventIds = append(eventIds, event.DbId)
	}

	eventTypes, err := handler.getTypeIdTypeMappingsFromDb(typeIds)
	if err != nil {
		return nil, deepError.New(fn, "get type id type mappings", err)
	}

	for _, event := range mapEvents {
		event.Type = eventTypes[event.Type.DbId]
	}

	eventTags, err := handler.getEventIdTagMappingsFromDb(eventIds)
	if err != nil {
		return nil, deepError.New(fn, "get event id tag mappings", err)
	}

	for eventId, tag := range eventTags {
		event := mapEvents[eventId]
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

func (handler *EventsHandler) getTypeIdTypeMappingsFromDb(typeIds []int) (map[int]*EventType, error) {
	// returns a map of Type id and Type
	fn := "getTypeIdTypeMappingsFromDb"
	lenIds := len(typeIds)

	if lenIds == 0 {
		return map[int]*EventType{}, nil
	}

	paramsS := ""
	if lenIds == 1 {
		paramsS = "?"
	} else {
		paramsS = "?" + strings.Repeat(", ?", lenIds-1)
	}

	query := strings.Replace(GET_EVENT_TYPE_QUERY, "<params>", paramsS, 1)

	params := make([]interface{}, lenIds)
	for i, typeId := range typeIds {
		params[i] = typeId
	}

	rows, err := handler.db.Query(query, params...)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	mapTypeIdType := make(map[int]*EventType, 0)
	for rows.Next() {
		eventType := &EventType{}
		rows.Scan(&eventType.DbId, &eventType.Value, &eventType.CreatedAt, &eventType.UpdatedAt, &eventType.Status)

		mapTypeIdType[eventType.DbId] = eventType
	}

	return mapTypeIdType, nil
}

func (handler *EventsHandler) getEventIdTagMappingsFromDb(eventIds []int) (map[int]*EventTag, error) {
	// returns a map of EventId and Event Tag
	fn := "getEventIdTagMappingsFromDb"
	lenIds := len(eventIds)

	if lenIds == 0 {
		return map[int]*EventTag{}, nil
	}

	paramsS := ""
	if lenIds == 1 {
		paramsS = "?"
	} else {
		paramsS = "?" + strings.Repeat(", ?", lenIds-1)
	}

	query := strings.Replace(GET_EVENT_TAGS_QUERY, "<params>", paramsS, 1)

	params := make([]interface{}, lenIds)
	for i, eventId := range eventIds {
		params[i] = eventId
	}

	rows, err := handler.db.Query(query, params...)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	mapEventIdTag := make(map[int]*EventTag, 0)
	for rows.Next() {
		eventTag := &EventTag{}
		var eventId int
		rows.Scan(&eventTag.DbId, &eventTag.Value, &eventTag.CreatedAt, &eventTag.UpdatedAt, &eventTag.Status, &eventId)

		mapEventIdTag[eventId] = eventTag
	}

	return mapEventIdTag, nil
}
