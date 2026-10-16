package inapi_events

import (
	"time"

	"github.com/gethiox/EDCITR/internal/inara/inapi"
)

const INAPI_URL = "https://inara.cz/inapi/v1/"

type InaraAPIEvent struct {
	// Name of the event/method
	EventName string `json:"eventName"`

	// Date and time in ISO 8601 format (like: 2017-05-02T17:30:49Z).`
	// Always use a correct date and time related to the event (from the journals),
	// for example when commander achieved a new rank, when was ship bought and so on.
	// If there is no date/time directly related to the event, just use the actual time.
	// The date/time provided shouldn't be older than 30 days.
	EventTimestamp time.Time `json:"eventTimestamp,format:'2006-01-02T15:04:05-0700'"`
	// To emit "Z" on zero-offset timezone, use '2006-01-02T15:04:05Z0700'

	// Optional custom event ID by your choice, used for the output (like: 1234)
	// EventCustomID  int

	// Required event properties/data.
	// See more details about individual event properties in the events documentation.
	EventData any `json:"eventData"`
}

type InaraAPIRequest struct {
	Header inapi.APIHeader `json:"header"`
	Events []InaraAPIEvent `json:"events"`
}
