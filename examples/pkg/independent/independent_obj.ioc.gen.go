// Code generated by "go-ioc"; DO NOT EDIT.
package independent

import (
	goIoc0 "github.com/catmorte/go-ioc/pkg/context"
)

func init() {
	dep0_0 := goIoc0.DepScoped[string]("config")
	goIoc0.RegScoped("independentScope1", func() *IndependentObj1 {
		v := &IndependentObj1{
			SomeDepField: goIoc0.ResolveDep[string](dep0_0),
		}
		v.Init()
		return v
	}, dep0_0)
	goIoc0.RegScoped("independentScope2", func() IndependentObj2 {
		v := IndependentObj2{}
		v.Init()
		return v
	})
	goIoc0.RegScoped("", func() independentObj3 {
		v := independentObj3{}
		v.Init()
		return v
	})

}
