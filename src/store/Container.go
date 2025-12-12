package Store

import (
	"fmt"
	"strings"
	"time"
)

type Object struct {
	Name                string    `json:"name"`
	Items               []Item    `json:"items"`
	Description         string    `json:"description"`
	Container           string    `json:"container"`
	Tags                []string  `json:"tags"`
	CreationDate        time.Time `json:"creationDate"`
	AvailableParameters []string  `json:"availableParameters"`
}

func (o Object) String() string {
	var desc string
	var container string
	if o.Description == "" {
		desc = "''"
	} else {
		desc = o.Description
	}

	if o.Container == "" {
		container = "''"
	} else {
		container = o.Container
	}
	return fmt.Sprintf("Name: %s \n Description: %s \n Container: %s \n Created: %v \n Tags: %v \n Items:",
		o.Name, desc, container, o.CreationDate, "["+strings.Join(o.Tags, ",")+"]") +
		func(i []Item) string {
			if i != nil && len(i) == 0 {
				return "''"
			}
			if i == nil {
				return "<nil>"
			}

			var items string
			for _, item := range i {
				items += "\n" + item.String()
			}
			return items
		}(o.Items)
}
