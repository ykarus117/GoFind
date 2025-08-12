package Object

import (
	"GoSafe/src/Item"
)

type Object struct {
	Name        string      `json:"name"`
	Items       []Item.Item `json:"items"`
	Description string      `json:"description"`
	Container   string      `json:"container"`
	Tags        []string    `json:"tags"`
	OwnerID     int
}
