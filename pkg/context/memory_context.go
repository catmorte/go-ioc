package context

import (
	"sync"
)

type memoryContext struct {
	storage  map[string]map[any]interface{}
	requests map[string]map[any][]chan interface{}
	lock     *sync.RWMutex
}

func (m *memoryContext) GetUnresolvedRequests() []*dependencyRequest {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var unresolvedRequests []*dependencyRequest
	for scope, scopes := range m.requests {
		for t, waiters := range scopes {
			for _, waiter := range waiters {
				unresolvedRequests = append(unresolvedRequests, &dependencyRequest{t, waiter, scope})
			}
		}
	}
	return unresolvedRequests
}

func (m *memoryContext) Ask(t any) chan interface{} {
	return m.AskScoped(DefaultScope, t)
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
	}()
	for _, r := range requests {
		if foundScope, ok := m.storage[r.Scope]; ok {
			if found, ok := foundScope[r.Type]; ok {
				r.Waiter <- found
				continue
			}
		}
		m.appendWaiter(r.Scope, r.Type, r.Waiter)
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
		storage:  map[string]map[any]interface{}{},
		requests: map[string]map[any][]chan interface{}{},
		lock:     &sync.RWMutex{}}
}
