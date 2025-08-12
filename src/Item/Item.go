package Item

type Item struct {
	Name        string   `json:"name"`
	Quantity    int      `json:"quantity"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Container   string   `json:"container"`
}
