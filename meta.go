package main

const (
	DB_ID      = "_id"
	CREATED_AT = "created_at"
	UPDATED_AT = "updated_at"
	STATUS     = "status"
)

const (
	EVENTS_TABLE_NAME    = "events"
	EVENTS_COL_ID        = "id"
	EVENTS_COL_TITLE     = "title"
	EVENTS_COL_NOTE      = "note"
	EVENTS_COL_TIMESTAMP = "timestamp"
	EVENTS_COL_TYPE_ID   = "type_id"
)

const (
	EVENT_TYPES_TABLE_NAME = "event_types"
	EVENT_TYPES_COL_VALUE  = "value"
)

const (
	EVENT_TAGS_TABLE_NAME = "event_tags"
	EVENT_TAGS_COL_VALUE  = "value"
)

const (
	EVENT_TAG_MAPPINGS_TABLE_NAME   = "event_tag_mappings"
	EVENT_TAG_MAPPINGS_COL_EVENT_ID = "event_id"
	EVENT_TAG_MAPPINGS_COL_TAG_ID   = "tag_id"
)
