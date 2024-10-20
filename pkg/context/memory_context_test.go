package context

import (
	"fmt"
	"testing"
)

func init() {
	SetContext(NewMemoryContext())
}

type testInterface interface {
	TestFunc() string
}
type firstIndependentStruct struct {
	val string
}
type secondIndependentStruct struct {
	val string
}
type dependentStruct struct {
	firstDep  *firstIndependentStruct
	secondDep *secondIndependentStruct
}

func (t *dependentStruct) TestFunc() string {
	return fmt.Sprintf("%v %v", t.firstDep.val, t.secondDep.val)
}

func TestMemoryContext_DefaultContext(t *testing.T) {
	Reg(func() *firstIndependentStruct {
		t.Log("Start init firstIndependentStruct")
		return &firstIndependentStruct{"firstTestString"}
	})

	firstDep := Dep[*firstIndependentStruct]()
	secondDep := Dep[*secondIndependentStruct]()
	Reg(func() *dependentStruct {
		t.Log("Start init dependentStruct")
		return &dependentStruct{
			firstDep:  (<-firstDep.Waiter).(*firstIndependentStruct),
			secondDep: (<-secondDep.Waiter).(*secondIndependentStruct),
		}
	}, firstDep, secondDep)

	Reg(func() *secondIndependentStruct {
		t.Log("Start init secondIndependentStruct")
		return &secondIndependentStruct{"secondTestString"}
	})

	t.Log("Start waiting for dependentStruct")
	actualInst := Ask[*dependentStruct]()
	if actualInst.firstDep.val == "firstTestString" && actualInst.secondDep.val == "secondTestString" {
		t.Log("Initialized")
		return
	}
	t.Errorf("Expected values %v %v", "firstTestString", "secondTestString")
}

func TestMemoryContext_CustomContext(t *testing.T) {
	const customScopeName = "custom"
	RegScoped(customScopeName, func() *firstIndependentStruct {
		t.Log("Start init firstIndependentStruct")
		return &firstIndependentStruct{"firstTestString"}
	})

	firstDep := DepScoped[*firstIndependentStruct](customScopeName)
	secondDep := DepScoped[*secondIndependentStruct](customScopeName)
	RegScoped(customScopeName, func() *dependentStruct {
		t.Log("Start init dependentStruct")
		return &dependentStruct{
			firstDep:  (<-firstDep.Waiter).(*firstIndependentStruct),
			secondDep: (<-secondDep.Waiter).(*secondIndependentStruct),
		}
	}, firstDep, secondDep)

	RegScoped(customScopeName, func() *secondIndependentStruct {
		t.Log("Start init secondIndependentStruct")
		return &secondIndependentStruct{"secondTestString"}
	})

	t.Log("Start waiting for dependentStruct")

	actualInst := AskScoped[*dependentStruct](customScopeName)
	if actualInst.firstDep.val == "firstTestString" && actualInst.secondDep.val == "secondTestString" {
		t.Log("Initialized")
		return
	}
	t.Errorf("Expected values %v %v", "firstTestString", "secondTestString")
}

func TestMemoryContext_DefaultContext_DuckTyping(t *testing.T) {
	Reg(func() *firstIndependentStruct {
		t.Log("Start init firstIndependentStruct")
		return &firstIndependentStruct{"firstTestString"}
	})

	firstDep := Dep[*firstIndependentStruct]()
	secondDep := Dep[*secondIndependentStruct]()
	Reg(func() *dependentStruct {
		t.Log("Start init dependentStruct")
		return &dependentStruct{
			firstDep:  (<-firstDep.Waiter).(*firstIndependentStruct),
			secondDep: (<-secondDep.Waiter).(*secondIndependentStruct),
		}
	}, firstDep, secondDep)

	Reg(func() *secondIndependentStruct {
		t.Log("Start init secondIndependentStruct")
		return &secondIndependentStruct{"secondTestString"}
	})

	t.Log("Start waiting for dependentStruct")
	interfaceVal := AskInterface[testInterface]()
	result := interfaceVal.TestFunc()

	if result == "firstTestString secondTestString" {
		t.Log("Initialized")
		return
	}
	t.Errorf("Expected value '%v'", "firstTestString secondTestString")
}

func TestMemoryContext_CustomContext_DuckTyping(t *testing.T) {
	const customScopeName = "custom"
	RegScoped(customScopeName, func() *firstIndependentStruct {
		t.Log("Start init firstIndependentStruct")
		return &firstIndependentStruct{"firstTestString"}
	})

	firstDep := DepScoped[*firstIndependentStruct](customScopeName)
	secondDep := DepScoped[*secondIndependentStruct](customScopeName)
	RegScoped(customScopeName, func() *dependentStruct {
		t.Log("Start init dependentStruct")
		return &dependentStruct{
			firstDep:  (<-firstDep.Waiter).(*firstIndependentStruct),
			secondDep: (<-secondDep.Waiter).(*secondIndependentStruct),
		}
	}, firstDep, secondDep)

	RegScoped(customScopeName, func() *secondIndependentStruct {
		t.Log("Start init secondIndependentStruct")
		return &secondIndependentStruct{"secondTestString"}
	})

	t.Log("Start waiting for dependentStruct")

	interfaceVal := AskInterfaceScoped[testInterface](customScopeName)
	result := interfaceVal.TestFunc()

	if result == "firstTestString secondTestString" {
		t.Log("Initialized")
		return
	}
	t.Errorf("Expected value '%v'", "firstTestString secondTestString")
}
