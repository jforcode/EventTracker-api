package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jforcode/Go-Util"
	"github.com/magiconair/properties"
)

func InitDb() *sql.DB {
	p := properties.MustLoadFile("test.properties", properties.UTF8)
	db, err := getDbFromProps(p)
	if err != nil {
		panic(err)
	}

	return db
}

func TestGetEventDao(t *testing.T) {
	db := InitDb()

	err := util.Db.ClearTables(db, eventTagMapTableName, eventsTableName, eventTagsTableName)
	HandleTestError(t, err)

	handler := &EventsHandler{}
	handler.Init(db)

	eventType, err := handler.dbStuff.findEventTypeByValue("start")
	HandleTestError(t, err)

	now := time.Now()
	event := &Event{
		ID:        "TestEvent",
		Title:     "Some Test Event",
		Note:      "Some Test Note",
		Timestamp: now,
		Type: &EventType{
			DbRecord: DbRecord{DbID: eventType.DbID},
		},
	}

	eventDbID, err := handler.dbStuff.insertEvent(event)
	HandleTestError(t, err)

	tag1 := &EventTag{Value: "tag1"}
	tag1ID, err := handler.dbStuff.insertEventTag(tag1)
	HandleTestError(t, err)

	tag2 := &EventTag{Value: "tag2"}
	tag2ID, err := handler.dbStuff.insertEventTag(tag2)
	HandleTestError(t, err)

	tag1Map := &EventTagMap{EventID: eventDbID, TagID: tag1ID}
	_, err = handler.dbStuff.insertEventTagMapping(tag1Map)
	HandleTestError(t, err)

	tag2Map := &EventTagMap{EventID: eventDbID, TagID: tag2ID}
	_, err = handler.dbStuff.insertEventTagMapping(tag2Map)
	HandleTestError(t, err)

	expectedEvent := &Event{
		DbRecord: DbRecord{DbID: eventDbID, Status: statusActive},
		ID:       "TestEvent",
		Title:    "Some Test Event",
		Note:     "Some Test Note",
		Type: &EventType{
			DbRecord: DbRecord{DbID: eventType.DbID, Status: statusActive},
			Value:    "start",
		},
		Timestamp: now,
		Tags: []*EventTag{
			&EventTag{
				DbRecord: DbRecord{DbID: tag1ID, Status: statusActive},
				Value:    "tag1",
			},
			&EventTag{
				DbRecord: DbRecord{DbID: tag2ID, Status: statusActive},
				Value:    "tag2",
			},
		},
	}

	actualEvent, err := handler.GetEvent("TestEvent")
	HandleTestError(t, err)
	if actualEvent == nil || actualEvent.Type == nil || actualEvent.Tags == nil || len(actualEvent.Tags) != len(expectedEvent.Tags) {
		HandleTestError(t, errors.New("Got invalid event data"))
	}

	actualEvent.CreatedAt = expectedEvent.CreatedAt
	actualEvent.UpdatedAt = expectedEvent.UpdatedAt
	actualEvent.Type.CreatedAt = expectedEvent.Type.CreatedAt
	actualEvent.Type.CreatedAt = expectedEvent.Type.CreatedAt
	for index, expTag := range expectedEvent.Tags {
		actTag := actualEvent.Tags[index]
		actTag.CreatedAt = expTag.CreatedAt
		actTag.UpdatedAt = expTag.UpdatedAt
	}

	if !cmp.Equal(actualEvent, expectedEvent) {
		HandleTestError(t, errors.New("Actual event and expected events don't match"))
	}

	err = util.Db.ClearTables(db, eventTagMapTableName, eventsTableName, eventTagsTableName)
	HandleTestError(t, err)
}

func TestCreateEventDao(t *testing.T) {
	db := InitDb()

	err := util.Db.ClearTables(db, eventTagMapTableName, eventsTableName, eventTagsTableName)
	HandleTestError(t, err)

	handler := &EventsHandler{}
	handler.Init(db)

	now := time.Now()
	event := &Event{
		Title: "Some Test Event",
		Note:  "Some Test Note",
		Type: &EventType{
			Value: "start",
		},
		Timestamp: now,
		Tags: []*EventTag{
			&EventTag{Value: "tag1"},
			&EventTag{Value: "tag2"},
		},
	}

	eventID, err := handler.CreateEvent(event)
	HandleTestError(t, err)

	eventType, err1 := handler.dbStuff.findEventTypeByValue("start")
	HandleTestError(t, err1)
	if eventType == nil {
		HandleTestError(t, errors.New("Event Type not found"))
	}

	queryEvents := fmt.Sprintf(
		"SELECT %s, %s, %s, %s, %s FROM %s",
		eventsColID, eventsColTitle, eventsColNote, eventsColTypeID, eventsColCreatedAt, eventsTableName,
	)
	typeIDStr := strconv.Itoa(int(eventType.DbID))
	equalEvents, err := util.Db.CompareDbData(db, queryEvents, make([]interface{}, 0), 5, [][]string{
		{eventID, "Some Test Event", "Some Test Note", typeIDStr, now.UTC().Format(time.RFC3339Nano)},
	}, false)
	HandleTestError(t, err)
	if !equalEvents {
		HandleTestError(t, errors.New("Invalid event data"))
	}

	queryTags := fmt.Sprintf(
		"SELECT %s FROM %s",
		eventTagsColValue, eventTagsTableName,
	)
	equalTags, err := util.Db.CompareDbData(db, queryTags, make([]interface{}, 0), 1, [][]string{
		{"tag1"},
		{"tag2"},
	}, false)
	HandleTestError(t, err)
	if !equalTags {
		HandleTestError(t, errors.New("Invalid tags data"))
	}

	dbEvent, err := handler.dbStuff.findEventByID(eventID)
	HandleTestError(t, err)
	tag1, err := handler.dbStuff.findEventTagByValue("tag1")
	HandleTestError(t, err)
	tag2, err := handler.dbStuff.findEventTagByValue("tag2")
	HandleTestError(t, err)

	queryTagMaps := fmt.Sprintf(
		"SELECT %s, %s FROM %s",
		eventTagMapColEventID, eventTagMapColTagID, eventTagMapTableName,
	)

	eventIDStr := strconv.Itoa(int(dbEvent.DbID))
	tag1IDStr := strconv.Itoa(int(tag1.DbID))
	tag2IDStr := strconv.Itoa(int(tag2.DbID))

	equalTagMaps, err := util.Db.CompareDbData(db, queryTagMaps, make([]interface{}, 0), 2, [][]string{
		{eventIDStr, tag1IDStr},
		{eventIDStr, tag2IDStr},
	}, false)
	HandleTestError(t, err)
	if !equalTagMaps {
		HandleTestError(t, errors.New("Invalid tag event maps data"))
	}

	err = util.Db.ClearTables(db, eventTagMapTableName, eventsTableName, eventTagsTableName)
	HandleTestError(t, err)
}
