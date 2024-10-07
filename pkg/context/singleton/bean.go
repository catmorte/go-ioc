package singleton

type (
	Bean[Out any] struct{}
)

func (Bean[Out]) isSingleton() {}

func (Bean[Out]) Init() {}
