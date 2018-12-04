package main

import (
	"database/sql"
	"testing"

	"github.com/jforcode/Util"
	"github.com/magiconair/properties"
)

func InitDb() *sql.DB {
	p := properties.MustLoadFile("test.properties", properties.UTF8)
	db, err := GetDbFromProperties(p)
	if err != nil {
		panic(err)
	}

	return db
}

func TestCreateEventDao(t *testing.T) {
	db := InitDb()
	util.Db.ClearTables(db, EVENTS_TABLE_NAME, EVENT_TAGS_TABLE_NAME, EVENT_TYPES_TABLE_NAME, EVENT_TAG_MAPPINGS_TABLE_NAME)

	handler := &EventsHandler{db}

	event := &Event{
		Title: "Some Test Event",
		Note:  "Some Test Note",
		Type: &EventType{
			Value: "start",
		},
		Tags: []*EventTag{
			&EventTag{Value: "tag1"},
		},
	}
	eventId, err := handler.CreateEvent(event)

	actualEvents, err := handler.GetAllEvents()

}
