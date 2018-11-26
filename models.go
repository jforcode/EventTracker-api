package main

import (
	"time"
)

type EventType string

const (
	EVENT_TYPE_START = "start"
	EVENT_TYPE_END   = "end"
)

type EventTag string

type DbRecord struct {
	DbId      int
	CreatedAt time.Time
	UpdatedAt time.Time
	Status    time.Time
}

type Event struct {
	DbRecord
	Id        string    `json:"id"`
	Title     string    `json:"title"`
	Note      string    `json:"note"`
	Timestamp time.Time `json:timestamp`
	Type      string    `json:type`
	Tags      []string  `json:tags`
}
