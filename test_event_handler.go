package main

import (
	"errors"
	"strconv"
	"strings"
)

type TestEventHandler struct {
	events      []*Event
	lastEventId int
}

func (handler *TestEventHandler) GetAllEvents() ([]*Event, error) {
	return handler.events, nil
}

func (handler *TestEventHandler) GetEvent(eventId string) (*Event, error) {
	for _, evt := range handler.events {
		if strings.EqualFold(eventId, evt.Id) {
			return evt, nil
		}
	}

	return nil, errors.New("Event with id " + eventId + " not found")
}

func (handler *TestEventHandler) CreateEvent(evt *Event) (string, error) {
	handler.lastEventId++
	evt.Id = strconv.Itoa(handler.lastEventId)
	handler.events = append(handler.events, evt)
	return evt.Id, nil
}
