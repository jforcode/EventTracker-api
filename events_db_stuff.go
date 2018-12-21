package main

import (
	"database/sql"
	"fmt"

	"github.com/jforcode/Go-DeepError"
	"github.com/jforcode/Go-Util"
)

// TODO: implement cache
// TODO: reduce code by commoning out

type dbStuff struct {
	db *sql.DB
}

func (dbStuff *dbStuff) findEventByID(eventID string) (*Event, error) {
	fn := "findEventById"

	query := fmt.Sprintf(`
		SELECT E.%s, E.%s, E.%s, E.%s, E.%s, E.%s, E.%s, E.%s, E.%s
		FROM %s E
		WHERE E.%s = ?`,
		colDbID, eventsColID, eventsColTitle, eventsColNote, eventsColCreatedAt, eventsColTypeID, colCreatedAt, colUpdatedAt, colStatus,
		eventsTableName,
		eventsColID)

	rows, err := dbStuff.db.Query(query, eventID)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	if rows.Next() {
		event := &Event{Type: &EventType{}}
		rows.Scan(&event.DbID, &event.ID, &event.Title, &event.Note, &event.UserCreatedAt, &event.Type.DbID, &event.CreatedAt, &event.UpdatedAt, &event.Status)
		return event, nil
	}

	return nil, nil
}

func (dbStuff *dbStuff) findEventTypeByValue(value string) (*EventType, error) {
	fn := "findEventTypeByValue"

	query := fmt.Sprintf(`
		SELECT ETP.%s, ETP.%s, ETP.%s, ETP.%s, ETP.%s
		FROM %s ETP
		WHERE ETP.%s = ?`,
		colDbID, eventTypesColValue, colCreatedAt, colUpdatedAt, colStatus,
		eventTypesTableName,
		eventTypesColValue)

	rows, err := dbStuff.db.Query(query, value)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	if rows.Next() {
		eventType := &EventType{}
		rows.Scan(&eventType.DbID, &eventType.Value, &eventType.CreatedAt, &eventType.UpdatedAt, &eventType.Status)

		return eventType, nil
	}

	return nil, nil
}

func (dbStuff *dbStuff) findEventTypeByID(id int64) (*EventType, error) {
	fn := "findEventTypeByID"

	query := fmt.Sprintf(`
		SELECT ETP.%s, ETP.%s, ETP.%s, ETP.%s, ETP.%s
		FROM %s ETP
		WHERE ETP.%s = ?`,
		colDbID, eventTypesColValue, colCreatedAt, colUpdatedAt, colStatus,
		eventTypesTableName,
		colDbID)

	rows, err := dbStuff.db.Query(query, id)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	if rows.Next() {
		eventType := &EventType{}
		rows.Scan(&eventType.DbID, &eventType.Value, &eventType.CreatedAt, &eventType.UpdatedAt, &eventType.Status)

		return eventType, nil
	}

	return nil, nil
}

func (dbStuff *dbStuff) findEventTagByValue(value string) (*EventTag, error) {
	fn := "findEventTagByValue"

	query := fmt.Sprintf(`
		SELECT ETG.%s, ETG.%s, ETG.%s, ETG.%s, ETG.%s
		FROM %s ETG
		WHERE ETG.%s = ?`,
		colDbID, eventTagsColValue, colCreatedAt, colUpdatedAt, colStatus,
		eventTagsTableName,
		eventTagsColValue)

	rows, err := dbStuff.db.Query(query, value)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	if rows.Next() {
		eventTag := &EventTag{}
		rows.Scan(&eventTag.DbID, &eventTag.Value, &eventTag.CreatedAt, &eventTag.UpdatedAt, &eventTag.Status)

		return eventTag, nil
	}

	return nil, nil
}

func (dbStuff *dbStuff) findEventTagByID(id int64) (*EventTag, error) {
	fn := "findEventTagByID"

	query := fmt.Sprintf(`
		SELECT ETG.%s, ETG.%s, ETG.%s, ETG.%s, ETG.%s
		FROM %s ETG
		WHERE ETG.%s = ?`,
		colDbID, eventTagsColValue, colCreatedAt, colUpdatedAt, colStatus,
		eventTagsTableName,
		colDbID)

	rows, err := dbStuff.db.Query(query, id)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	if rows.Next() {
		eventTag := &EventTag{}
		rows.Scan(&eventTag.DbID, &eventTag.Value, &eventTag.CreatedAt, &eventTag.UpdatedAt, &eventTag.Status)

		return eventTag, nil
	}

	return nil, nil
}

func (dbStuff *dbStuff) findEventTagsByEventID(eventID string) ([]*EventTag, error) {
	fn := "findEventTagsByEventID"

	query := fmt.Sprintf(`
		SELECT ETG.%s, ETG.%s, ETG.%s, ETG.%s, ETG.%s
		FROM %s ETG
		JOIN %s ETM ON ETM.%s = ETG.%s
		JOIN %s E ON E.%s = ETM.%s AND E.%s = ?`,
		colDbID, eventTagsColValue, colCreatedAt, colUpdatedAt, colStatus,
		eventTagsTableName,
		eventTagMapTableName, eventTagMapColTagID, colDbID,
		eventsTableName, colDbID, eventTagMapColEventID, eventsColID)

	rows, err := dbStuff.db.Query(query, eventID)
	if err != nil {
		return nil, deepError.New(fn, "query", err)
	}
	defer rows.Close()

	eventTags := make([]*EventTag, 0)
	for rows.Next() {
		eventTag := &EventTag{}
		rows.Scan(&eventTag.DbID, &eventTag.Value, &eventTag.CreatedAt, &eventTag.UpdatedAt, &eventTag.Status)
		eventTags = append(eventTags, eventTag)
	}

	return eventTags, nil
}

func (dbStuff *dbStuff) insertEvent(event *Event) (int64, error) {
	fn := "insertEvent"

	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?)",
		eventsTableName, eventsColID, eventsColTitle, eventsColNote, eventsColTypeID, eventsColCreatedAt)

	res, err := util.Db.PrepareAndExec(dbStuff.db, query, event.ID, event.Title, event.Note, event.Type.DbID, event.UserCreatedAt)
	if err != nil {
		return -1, deepError.New(fn, "prepare and exec", err)
	}

	return getDbID(res)
}

func (dbStuff *dbStuff) insertEventType(eventType *EventType) (int64, error) {
	fn := "insertEventType"

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (?)",
		eventTypesTableName, eventTypesColValue)

	res, err := util.Db.PrepareAndExec(dbStuff.db, query, eventType.Value)
	if err != nil {
		return -1, deepError.New(fn, "prepare and exec", err)
	}

	return getDbID(res)

}

func (dbStuff *dbStuff) insertEventTag(eventTag *EventTag) (int64, error) {
	fn := "insertEventTag"

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (?)",
		eventTagsTableName, eventTagsColValue)

	res, err := util.Db.PrepareAndExec(dbStuff.db, query, eventTag.Value)
	if err != nil {
		return -1, deepError.New(fn, "prepare and exec", err)
	}

	return getDbID(res)

}

func (dbStuff *dbStuff) insertEventTagMapping(eventTagMap *EventTagMap) (int64, error) {
	fn := "insertEventTagMapping"

	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES (?, ?)",
		eventTagMapTableName, eventTagMapColEventID, eventTagMapColTagID)

	res, err := util.Db.PrepareAndExec(dbStuff.db, query, eventTagMap.EventID, eventTagMap.TagID)
	if err != nil {
		return -1, deepError.New(fn, "prepare and exec", err)
	}

	return getDbID(res)
}

func getDbID(res sql.Result) (int64, error) {
	fn := "getDbID"

	lastID, err := res.LastInsertId()
	if err != nil {
		return -1, deepError.New(fn, "last insert id", err)
	}

	return lastID, nil
}
