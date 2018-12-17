package main

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jforcode/Go-DeepError"
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

func TestCreateEventDao(t *testing.T) {
	db := InitDb()

	// setup
	err := util.Db.ClearTables(db, eventsTableName, eventTagsTableName, eventTagMapTableName)
	HandleTestError(t, err)

	handler := &EventsHandler{db}

	event := &Event{
		Title: "Some Test Event",
		Note:  "Some Test Note",
		Type: &EventType{
			Value: "start",
		},
		Tags: []*EventTag{
			&EventTag{Value: "tag1"},
			&EventTag{Value: "tag2"},
		},
	}

	eventID, err := handler.CreateEvent(event)
	if err != nil {
		HandleTestError(t, err)
	}

	eventType, err1 := findEventTypeByValue("start")
	HandleTestError(t, err1)
	if eventType == nil {
		HandleTestError(t, errors.New("Event Type not found"))
	}

	query := fmt.Sprintf(
		"SELECT %s, %s, %s, %s FROM %s",
		eventsColID, eventsColTitle, eventsColNote, eventsColTypeID, eventsTableName,
	)
	AssertDbData(db, query, make([]interface{}, 0), 4, [][]interface{}{
		[]interface{}{eventID, "Some Test Event", "Some Test Note", eventType.DbID},
	})

	// assert tags are created
	// assert tag mappings are created

	util.Db.ClearTables(db, eventsTableName, eventTagsTableName, eventTagMapTableName)
}

func AssertDbData(db *sql.DB, query string, args []interface{}, numCols int, expected [][]interface{}) (bool, error) {
	fn := "AssertDbData"
	rows, err := db.Query(query, args...)
	if err != nil {
		return false, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	for rows.Next() {
		rowData := make([]interface{}, numCols)
		rows.Scan(rowData...)
		fmt.Printf("Got scan data: %+v", rowData)

		if !cmp.Equal(rowData, expected) {
			return false, nil
		}
	}

	return true, nil
}
