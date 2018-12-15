package main

import (
	"errors"
	"strconv"
	"strings"
)

// TestEventHandler is a dummy event handler which uses an array for its operations
type TestEventHandler struct {
	events      []*Event
	lastEventID int
}

// GetAllEvents gets all the events in the array
func (handler *TestEventHandler) GetAllEvents() ([]*Event, error) {
	return handler.events, nil
}

// GetEvent gets a specific event based on id from the array
func (handler *TestEventHandler) GetEvent(eventID string) (*Event, error) {
	for _, evt := range handler.events {
		if strings.EqualFold(eventID, evt.ID) {
			return evt, nil
		}
	}

	return nil, errors.New("Event with id " + eventID + " not found")
}

// CreateEvent adds a new event to the array
func (handler *TestEventHandler) CreateEvent(evt *Event) (string, error) {
	handler.lastEventID++
	evt.ID = strconv.Itoa(handler.lastEventID)
	handler.events = append(handler.events, evt)
	return evt.ID, nil
}
