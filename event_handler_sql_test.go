package main

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

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
	if err != nil {
		HandleTestError(t, err)
	}

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
	equalEvents, err := CompareDbData(db, queryEvents, make([]interface{}, 0), 5, [][]string{
		{eventID, "Some Test Event", "Some Test Note", typeIDStr, now.UTC().Format(time.RFC3339Nano)},
	}, true)
	HandleTestError(t, err)
	if !equalEvents {
		HandleTestError(t, errors.New("Invalid event data"))
	}

	queryTags := fmt.Sprintf(
		"SELECT %s FROM %s",
		eventTagsColValue, eventTagsTableName,
	)
	equalTags, err := CompareDbData(db, queryTags, make([]interface{}, 0), 1, [][]string{
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

	equalTagMaps, err := CompareDbData(db, queryTagMaps, make([]interface{}, 0), 2, [][]string{
		{eventIDStr, tag1IDStr},
		{eventIDStr, tag2IDStr},
	}, false)
	HandleTestError(t, err)
	if !equalTagMaps {
		HandleTestError(t, errors.New("Invalid tag event maps data"))
	}

	// util.Db.ClearTables(db, eventTagMapTableName, eventTagsTableName, eventsTableName)
}

func CompareDbData(db *sql.DB, query string, args []interface{}, numCols int, expected [][]string, debug bool) (bool, error) {
	fn := "CompareDbData"
	rows, err := db.Query(query, args...)
	if err != nil {
		return false, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	actual := make([][]string, 0)
	for rows.Next() {
		rowData := make([]string, numCols)
		rowDataPtr := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			rowDataPtr[i] = &rowData[i]
		}

		err := rows.Scan(rowDataPtr...)
		if err != nil {
			return false, deepError.New(fn, "scan", err)
		}

		actual = append(actual, rowData)
	}

	if debug {
		fmt.Println("\nActual")
		for _, act := range actual {
			for _, val := range act {
				fmt.Println(val, reflect.TypeOf(val))
			}
		}

		fmt.Println("\nExpected")
		for _, exp := range expected {
			for _, val := range exp {
				fmt.Println(val, reflect.TypeOf(val))
			}
		}
	}

	if len(actual) != len(expected) {
		return false, nil
	}

	for ind, act := range actual {
		if !cmp.Equal(act, expected[ind]) {
			return false, nil
		}
	}

	return true, nil
}
