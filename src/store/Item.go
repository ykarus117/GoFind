package Store

import (
	"fmt"
	"strings"
)

type Item struct {
	Id                  int      `json:"id"`
	Name                string   `json:"name"`
	Quantity            int      `json:"quantity"`
	Description         string   `json:"description"`
	Tags                []string `json:"tags"`
	Container           string   `json:"container"`
	AvailableParameters []string `json:"availableParameters"`
}

func (item Item) String() string {
	return fmt.Sprintf("\n -%s, id:%d \n\tQuantity: %d,\n\tContainer: %s\n\tTags: ", item.Name, item.Id, item.Quantity, item.Container) + "[" + strings.Join(item.Tags, ",") + "]\n"
}
