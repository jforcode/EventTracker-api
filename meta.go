package main

const (
	colDbID      = "_id"
	colCreatedAt = "_created_at"
	colUpdatedAt = "_updated_at"
	colStatus    = "_status"
)

const (
	eventsTableName    = "events"
	eventsColID        = "id"
	eventsColTitle     = "title"
	eventsColNote      = "note"
	eventsColCreatedAt = "created_at"
	eventsColTypeID    = "type_id"
)

const (
	eventTypesTableName = "event_types"
	eventTypesColValue  = "value"
)

const (
	eventTagsTableName = "event_tags"
	eventTagsColValue  = "value"
)

const (
	eventTagMapTableName  = "event_tag_mappings"
	eventTagMapColEventID = "event_id"
	eventTagMapColTagID   = "tag_id"
)
