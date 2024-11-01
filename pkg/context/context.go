package context

import (
	"sync"
)

const DefaultScope = ""

var (
	CurrentContext Context = NewMemoryContext()
	lock           sync.RWMutex
)

type (
	prototype[T any] struct {
		constructor func() T
	}

	dependencyRequest struct {
		Type        any
		Waiter      chan any
		Scope       string
		toInterface bool
	}

	Context interface {
		Reg(interfaceNil any, constructor func() interface{}, request ...*dependencyRequest)
		Ask(interfaceNil any) chan interface{}
		AskInterface(interfaceNil any) chan interface{}

		RegScoped(scope string, interfaceNil any, constructor func() interface{}, request ...*dependencyRequest)
		AskScoped(scope string, interfaceNil any) chan interface{}
		AskInterfaceScoped(scope string, interfaceNil any) chan interface{}

		GetUnresolvedRequests() []*dependencyRequest
	}
)

func DepInterface[T any]() *dependencyRequest {
	return &dependencyRequest{(*T)(nil), make(chan any, 1), DefaultScope, true}
}

func DepInterfaceScoped[T any](scope string) *dependencyRequest {
	return &dependencyRequest{(*T)(nil), make(chan any, 1), scope, true}
}

func Dep[T any]() *dependencyRequest {
	return &dependencyRequest{(*T)(nil), make(chan any, 1), DefaultScope, false}
}

func DepScoped[T any](scope string) *dependencyRequest {
	return &dependencyRequest{(*T)(nil), make(chan any, 1), scope, false}
}

func SetContext(context Context) {
	lock.Lock()
	defer lock.Unlock()
	CurrentContext = context
}

func GetContext() Context {
	lock.RLock()
	defer lock.RUnlock()
	return CurrentContext
}

func Ask[T any]() T {
	value := (<-GetContext().Ask((*T)(nil)))
	if prototype, ok := value.(*prototype[T]); ok {
		return prototype.constructor()
	}
	return value.(T)
}

func AskInterface[T any]() T {
	value := (<-GetContext().AskInterface((*T)(nil)))
	if prototype, ok := value.(*prototype[T]); ok {
		return prototype.constructor()
	}
	return value.(T)
}

func Reg[T any](constructor func() T, request ...*dependencyRequest) {
	GetContext().Reg((*T)(nil), typeToAnyFunc[T](constructor), request...)
}

func RegPrototype[T any](constructor func() T, request ...*dependencyRequest) {
	GetContext().Reg((*T)(nil), func() interface{} {
		return &prototype[T]{constructor: constructor}
	}, request...)
}

func AskScoped[T any](scope string) T {
	value := (<-GetContext().AskScoped(scope, (*T)(nil)))
	if prototype, ok := value.(*prototype[T]); ok {
		return prototype.constructor()
	}
	return value.(T)
}

func AskInterfaceScoped[T any](scope string) T {
	value := (<-GetContext().AskInterfaceScoped(scope, (*T)(nil)))
	if prototype, ok := value.(*prototype[T]); ok {
		return prototype.constructor()
	}
	return value.(T)
}

func RegScoped[T any](scope string, constructor func() T, request ...*dependencyRequest) {
	GetContext().RegScoped(scope, (*T)(nil), typeToAnyFunc[T](constructor), request...)
}

func RegPrototypeScoped[T any](scope string, constructor func() T, request ...*dependencyRequest) {
	GetContext().RegScoped(scope, (*T)(nil), func() interface{} {
		return &prototype[T]{constructor: constructor}
	}, request...)
}

func ResolveDep[T any](dep *dependencyRequest) T {
	rawVal := <-dep.Waiter
	go func() {
		dep.Waiter <- rawVal
	}()

	if prototype, ok := rawVal.(*prototype[T]); ok {
		return prototype.constructor()
	}
	return (rawVal).(T)
}

func typeToAnyFunc[T any](f func() T) func() any {
	return func() any {
		return f()
	}
}
