package main

import (
	"time"
)

const (
	EVENT_TYPE_START = "start"
	EVENT_TYPE_END   = "end"
)

const (
	STATUS_ACTIVE  = "active"
	STATUS_DELETED = "deleted"
)

type DbRecord struct {
	DbId      int
	CreatedAt time.Time
	UpdatedAt time.Time
	Status    string
}

type Event struct {
	DbRecord
	Id        string      `json:"id"`
	Title     string      `json:"title"`
	Note      string      `json:"note"`
	Timestamp time.Time   `json:timestamp`
	Type      *EventType  `json:type`
	Tags      []*EventTag `json:tags`
}

type EventType struct {
	DbRecord
	Value string `json:"value"`
}

type EventTag struct {
	DbRecord
	Value string `json:"value"`
}

type EventTagMap struct {
	DbRecord
	EventId string `json:"event_id"`
	TagId   string `json:"tag_id"`
}
