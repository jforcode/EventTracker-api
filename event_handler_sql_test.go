package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

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
	fn := "TestGetEventDao"
	db := InitDb()

	err := util.Db.ClearTables(db, eventTagMapTableName, eventsTableName, eventTagsTableName)
	util.Test.HandleIfTestError(t, err, fn)

	defer util.Db.ClearTables(db, eventTagMapTableName, eventsTableName, eventTagsTableName)

	handler := &EventsHandler{}
	handler.Init(db)

	eventType, err := handler.dbStuff.findEventTypeByValue("start")
	util.Test.HandleIfTestError(t, err, fn)

	now := time.Now()
	event := &Event{
		ID:            "TestEvent",
		Title:         "Some Test Event",
		Note:          "Some Test Note",
		UserCreatedAt: now,
		Type: &EventType{
			DbRecord: DbRecord{DbID: eventType.DbID},
		},
	}

	eventDbID, err := handler.dbStuff.insertEvent(event)
	util.Test.HandleIfTestError(t, err, fn)

	tag1 := &EventTag{Value: "tag1"}
	tag1ID, err := handler.dbStuff.insertEventTag(tag1)
	util.Test.HandleIfTestError(t, err, fn)

	tag2 := &EventTag{Value: "tag2"}
	tag2ID, err := handler.dbStuff.insertEventTag(tag2)
	util.Test.HandleIfTestError(t, err, fn)

	tag1Map := &EventTagMap{EventID: eventDbID, TagID: tag1ID}
	_, err = handler.dbStuff.insertEventTagMapping(tag1Map)
	util.Test.HandleIfTestError(t, err, fn)

	tag2Map := &EventTagMap{EventID: eventDbID, TagID: tag2ID}
	_, err = handler.dbStuff.insertEventTagMapping(tag2Map)
	util.Test.HandleIfTestError(t, err, fn)

	expectedEvent := &Event{
		DbRecord: DbRecord{DbID: eventDbID, Status: statusActive},
		ID:       "TestEvent",
		Title:    "Some Test Event",
		Note:     "Some Test Note",
		Type: &EventType{
			DbRecord: DbRecord{DbID: eventType.DbID, Status: statusActive},
			Value:    "start",
		},
		UserCreatedAt: now.UTC(),
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
	util.Test.HandleIfTestError(t, err, fn)
	if actualEvent == nil || actualEvent.Type == nil || actualEvent.Tags == nil || len(actualEvent.Tags) != len(expectedEvent.Tags) {
		util.Test.HandleIfTestError(t, errors.New("Got invalid event data"), fn)
	}

	actualEvent.CreatedAt = expectedEvent.CreatedAt
	actualEvent.UpdatedAt = expectedEvent.UpdatedAt
	actualEvent.Type.CreatedAt = expectedEvent.Type.CreatedAt
	actualEvent.Type.UpdatedAt = expectedEvent.Type.UpdatedAt
	for index, expTag := range expectedEvent.Tags {
		actTag := actualEvent.Tags[index]
		actTag.CreatedAt = expTag.CreatedAt
		actTag.UpdatedAt = expTag.UpdatedAt
	}

	util.Test.AssertEquals(t, expectedEvent, actualEvent, "Events don't match")
}

func TestCreateEventDao(t *testing.T) {
	fn := "TestCreateEventDao"
	db := InitDb()

	err := util.Db.ClearTables(db, eventTagMapTableName, eventsTableName, eventTagsTableName)
	util.Test.HandleIfTestError(t, err, fn)

	defer util.Db.ClearTables(db, eventTagMapTableName, eventsTableName, eventTagsTableName)

	handler := &EventsHandler{}
	handler.Init(db)

	now := time.Now()
	event := &Event{
		Title: "Some Test Event",
		Note:  "Some Test Note",
		Type: &EventType{
			Value: "start",
		},
		UserCreatedAt: now,
		Tags: []*EventTag{
			&EventTag{Value: "tag1"},
			&EventTag{Value: "tag2"},
		},
	}

	eventID, err := handler.CreateEvent(event)
	util.Test.HandleIfTestError(t, err, fn)

	eventType, err1 := handler.dbStuff.findEventTypeByValue("start")
	util.Test.HandleIfTestError(t, err1, fn)
	if eventType == nil {
		util.Test.HandleIfTestError(t, errors.New("Event Type not found"), fn)
	}

	queryEvents := fmt.Sprintf(
		"SELECT %s, %s, %s, %s, %s FROM %s",
		eventsColID, eventsColTitle, eventsColNote, eventsColTypeID, eventsColCreatedAt, eventsTableName,
	)
	typeIDStr := strconv.Itoa(int(eventType.DbID))
	equalEvents, err := util.Db.CompareDbData(db, queryEvents, make([]interface{}, 0), 5, [][]string{
		{eventID, "Some Test Event", "Some Test Note", typeIDStr, now.UTC().Format(time.RFC3339Nano)},
	}, false)
	util.Test.HandleIfTestError(t, err, fn)

	if !equalEvents {
		util.Test.HandleIfTestError(t, errors.New("Invalid Event Data"), fn)
	}

	queryTags := fmt.Sprintf(
		"SELECT %s FROM %s",
		eventTagsColValue, eventTagsTableName,
	)
	equalTags, err := util.Db.CompareDbData(db, queryTags, make([]interface{}, 0), 1, [][]string{
		{"tag1"},
		{"tag2"},
	}, false)
	util.Test.HandleIfTestError(t, err, fn)
	if !equalTags {
		util.Test.HandleIfTestError(t, errors.New("Invalid Tags Data"), fn)
	}

	dbEvent, err := handler.dbStuff.findEventByID(eventID)
	util.Test.HandleIfTestError(t, err, fn)
	tag1, err := handler.dbStuff.findEventTagByValue("tag1")
	util.Test.HandleIfTestError(t, err, fn)
	tag2, err := handler.dbStuff.findEventTagByValue("tag2")
	util.Test.HandleIfTestError(t, err, fn)

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
	util.Test.HandleIfTestError(t, err, fn)
	if !equalTagMaps {
		util.Test.HandleIfTestError(t, errors.New("Invalid Tag Map Data"), fn)
	}
}
