package main

import (
	"time"
)

const (
	eventTypeStart = "start"
	eventTypeEnd   = "end"
)

const (
	statusActive  = "active"
	statusDeleted = "deleted"
)

// DbRecord is the base model for a database struct
type DbRecord struct {
	DbID      int64     `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Status    string    `json:"-"`
}

// Event is the Db model to represent the event in a person's life.
type Event struct {
	DbRecord
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Note          string      `json:"note"`
	UserCreatedAt time.Time   `json:"created_at"`
	Type          *EventType  `json:"type"`
	Tags          []*EventTag `json:"tags"`
}

// EventType is the Db model for the type of the event. Can be start/end/distraction or anything else.
type EventType struct {
	DbRecord
	Value string `json:"value"`
}

// EventTag is the Db model for tags applied to an event
type EventTag struct {
	DbRecord
	Value string `json:"value"`
}

// EventTagMap is the Db model for the mapping between an event and a tag, as it is a m:n mapping
type EventTagMap struct {
	DbRecord
	EventID int64
	TagID   int64
}
