//go:generate go-ioc
package independent

import (
	singleton "github.com/catmorte/go-ioc/pkg/context/singleton"
)

type IndependentObj1 struct {
	singleton.Bean[*IndependentObj1] `bean:"independentScope1"`
	SomeDepField                     string `bean:"config"`
}

type IndependentObj2 struct {
	singleton.Bean[IndependentObj2] `bean:"independentScope2"`
	SimpleValue                     int
}

func (o *IndependentObj2) Init() {
	o.SimpleValue = 42
}
