package messages

import (
	"fmt"

	line "line-adaptor/internal/line"
)

// Location holds the parsed fields of a LINE location message.
type Location struct {
	Id        string
	Title     string  // optional
	Address   string  // optional
	Latitude  float64
	Longitude float64
}

// MessageType implements Parsed.
func (l *Location) MessageType() string { return "location" }

// ParseLocation parses a raw line.Message of type "location".
// Required fields: Id, Latitude, Longitude (both non-zero). Title and Address are optional.
func ParseLocation(msg *line.Message) (*Location, error) {
	if msg.Id == "" {
		return nil, fmt.Errorf("messages/location: Id is required")
	}
	if msg.Latitude == 0 {
		return nil, fmt.Errorf("messages/location: Latitude is required")
	}
	if msg.Longitude == 0 {
		return nil, fmt.Errorf("messages/location: Longitude is required")
	}
	return &Location{
		Id:        msg.Id,
		Title:     msg.Title,
		Address:   msg.Address,
		Latitude:  msg.Latitude,
		Longitude: msg.Longitude,
	}, nil
}
