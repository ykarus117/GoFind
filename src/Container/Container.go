package Container

import "GoSafe/src/Object"

type Container interface {
}

type ContainerImpl struct {
	name    string
	id      string
	isEmpty bool
	isFull  bool
	items   map[string]*Object.ItemObj
}
