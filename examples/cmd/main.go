package main

import (
	"fmt"

	"github.com/catmorte/go-ioc/examples/pkg/dependent"
	_ "github.com/catmorte/go-ioc/examples/pkg/independent"
	. "github.com/catmorte/go-ioc/pkg/context"
)

func init() {
	RegScoped("config", func() string {
		return "config_string"
	})
}

func main() {
	dependentBean1 := Ask[*dependent.DependentObj]()
	dependentBean2 := Ask[*dependent.DependentObj]()
	fmt.Println(dependentBean1.IndependentObj1.SomeDepField)
	fmt.Println(dependentBean2.IndependentObj2.SimpleValue)
}
