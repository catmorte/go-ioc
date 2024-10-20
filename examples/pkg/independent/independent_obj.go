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

type independentObj3 struct {
	singleton.Bean[independentObj3] `bean:""`
	SimpleValue                     int
}

func (i independentObj3) SomeSpecificLogicFunc() {
	println(i.SimpleValue)
}

func (i *independentObj3) Init() {
	i.SimpleValue = 69
}

func (o *IndependentObj2) Init() {
	o.SimpleValue = 42
}
