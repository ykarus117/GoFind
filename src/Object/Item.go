package Object

type Item interface {
	CreateItem(obj ItemObj) error
	DeleteItem() error
}

type ItemObj struct {
	Name        string
	Quantity    int
	Description string
}
