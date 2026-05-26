package event

import (
	"sync"
)

// Subscriber is a callback for event notifications.
type Subscriber func(eventType string, data any)

type eventHandler struct {
	fn   func(any)
	name string
}

// Emitter dispatches events to registered subscribers.
type Emitter struct {
	mu        sync.RWMutex
	subs      map[string][]eventHandler
	subFn     []Subscriber // generic subscribers
}

func NewEmitter() *Emitter {
	return &Emitter{
		subs: make(map[string][]eventHandler),
	}
}

// Subscribe registers a generic subscriber (callback receives eventType + data).
func (e *Emitter) Subscribe(fn Subscriber) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.subFn = append(e.subFn, fn)
}

// On registers a handler for a specific eventType (fn receives only data).
func (e *Emitter) On(eventType string, fn func(any)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.subs[eventType] = append(e.subs[eventType], eventHandler{fn: fn})
}

// Emit fires an event to all registered handlers.
func (e *Emitter) Emit(eventType string, data any) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Specific handlers
	for _, h := range e.subs[eventType] {
		go h.fn(data)
	}

	// Generic subscribers
	for _, fn := range e.subFn {
		go fn(eventType, data)
	}
}