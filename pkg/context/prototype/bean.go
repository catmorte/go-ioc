package singleton

type (
	Bean[Out any] struct{}
)

func (Bean[Out]) isPrototype() {}

func (Bean[Out]) Init() {}
