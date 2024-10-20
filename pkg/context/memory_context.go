package context

import (
	"reflect"
	"sync"
)

type memoryContext struct {
	storage           map[string]map[any]interface{}
	requests          map[string]map[any][]chan interface{}
	interfaceRequests map[string]map[any][]chan interface{}
	lock              *sync.RWMutex
}

func (m *memoryContext) GetUnresolvedRequests() []*dependencyRequest {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var unresolvedRequests []*dependencyRequest
	for scope, scopes := range m.requests {
		for t, waiters := range scopes {
			for _, waiter := range waiters {
				unresolvedRequests = append(unresolvedRequests, &dependencyRequest{t, waiter, scope, false})
			}
		}
	}
	return unresolvedRequests
}

func (m *memoryContext) Ask(t any) chan interface{} {
	return m.AskScoped(DefaultScope, t)
}

func (m *memoryContext) AskInterface(t any) chan interface{} {
	return m.AskInterfaceScoped(DefaultScope, t)
}

func (m *memoryContext) appendWaiter(s string, t any, waiter chan interface{}) {
	scope, ok := m.requests[s]
	if !ok {
		scope = map[any][]chan interface{}{}
		m.requests[s] = scope
	}

	typ, ok := scope[t]
	if !ok {
		typ = []chan any{}
		scope[t] = typ
	}
	scope[t] = append(typ, waiter)
}

func (m *memoryContext) appendInterfaceWaiter(s string, t any, waiter chan interface{}) {
	scope, ok := m.interfaceRequests[s]
	if !ok {
		scope = map[any][]chan interface{}{}
		m.interfaceRequests[s] = scope
	}

	typ, ok := scope[t]
	if !ok {
		typ = []chan any{}
		scope[t] = typ
	}
	scope[t] = append(typ, waiter)
}

func (m *memoryContext) Reg(t any, constructor func() interface{}, requests ...*dependencyRequest) {
	m.RegScoped(DefaultScope, t, constructor, requests...)
}

func (m *memoryContext) RegScoped(s string, t any, constructor func() interface{}, requests ...*dependencyRequest) {
	m.lock.Lock()
	defer m.lock.Unlock()

	go func() {
		instance := constructor()
		m.lock.Lock()
		defer m.lock.Unlock()
		scope, ok := m.storage[s]
		if !ok {
			scope = map[any]interface{}{}
			m.storage[s] = scope
		}

		scope[t] = instance
		m.notify(s, t, instance)
		m.notifyInterfaces(s, t, instance)
	}()
	for _, r := range requests {
		if foundScope, ok := m.storage[r.Scope]; ok {
			if found, ok := foundScope[r.Type]; ok {
				r.Waiter <- found
				continue
			}
		}
		if r.toInterface {
			m.appendInterfaceWaiter(r.Scope, r.Type, r.Waiter)
		} else {
			m.appendWaiter(r.Scope, r.Type, r.Waiter)
		}
	}
}

func (m *memoryContext) AskScoped(s string, t any) chan interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()

	waiter := make(chan any, 1)

	if foundScope, ok := m.storage[s]; ok {
		if found, ok := foundScope[t]; ok {
			waiter <- found
			return waiter
		}
	}

	m.appendWaiter(s, t, waiter)
	return waiter
}

func (m *memoryContext) AskInterfaceScoped(s string, t any) chan interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()

	waiter := make(chan any, 1)

	tType := reflect.TypeOf(t)
	if tType.Elem().Kind() == reflect.Interface {
		if foundScope, ok := m.storage[s]; ok {
			for v, val := range foundScope {
				valueType := reflect.TypeOf(v)
				ifaceType := tType.Elem()
				if valueType.Elem().Implements(ifaceType) {
					waiter <- val
					return waiter
				}
			}
		}
	} else {
		panic("unexpected type, interface type expected")
	}

	m.appendInterfaceWaiter(s, t, waiter)
	return waiter
}

func (m *memoryContext) notifyInterfaces(s string, valueTypeValue any, value interface{}) {
	if scope, ok := m.interfaceRequests[s]; ok {
		valueType := reflect.TypeOf(valueTypeValue)
		for t, waiters := range scope {
			tType := reflect.TypeOf(t)
			ifaceType := tType.Elem()
			if valueType.Elem().Implements(ifaceType) {
				for _, w := range waiters {
					w <- value
				}
				delete(scope, t)
				return
			}
		}
	}
}

func (m *memoryContext) notify(s string, t any, value interface{}) {
	if scope, ok := m.requests[s]; ok {
		if waiters, ok := scope[t]; ok {
			for _, w := range waiters {
				w <- value
			}
			delete(scope, t)
		}
	}
}

func NewMemoryContext() Context {
	return &memoryContext{
		storage:           map[string]map[any]interface{}{},
		requests:          map[string]map[any][]chan interface{}{},
		interfaceRequests: map[string]map[any][]chan interface{}{},
		lock:              &sync.RWMutex{},
	}
}
