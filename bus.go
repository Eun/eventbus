package eventbus

import (
	"container/ring"
	"sync"
)

const defaultHistorySize = 10
const defaultChannelSize = 10

// EventBus is a multi consumer / multi producer bus
type EventBus struct {
	listeners []chan Event
	wg        sync.WaitGroup
	mu        sync.Mutex
	events    *ring.Ring
	closed    bool
}

// New creates a new EventBus
func New(historySize ...int) *EventBus {
	size := defaultHistorySize
	if len(historySize) >= 1 {
		size = historySize[0]
	}

	bus := EventBus{}
	if size > 0 {
		bus.events = ring.New(size)
	}

	return &bus
}

// HistorySize returns the events that will be stored for future channels
func (bus *EventBus) HistorySize() int {
	if bus.events == nil {
		return 0
	}
	return bus.events.Len()
}

// AddListener adds an listener to the bus
// all cached messages will be immediately passed to the listener
func (bus *EventBus) AddListener(channelSize ...int) <-chan Event {
	size := defaultChannelSize
	if len(channelSize) >= 1 {
		size = channelSize[0]
	}
	var ch chan Event
	bus.mu.Lock()
	if !bus.closed {
		bus.wg.Add(1)
		ch = make(chan Event, size)
		bus.listeners = append(bus.listeners, ch)
		if bus.events != nil {
			bus.events.Do(func(p interface{}) {
				if p != nil {
					ch <- p.(Event)
				}
			})
		}
	}
	bus.mu.Unlock()
	return ch
}

// RemoveListener removes an listener from the bus
func (bus *EventBus) RemoveListener(listeners ...<-chan Event) {
	bus.mu.Lock()
	if !bus.closed {
		for _, ch := range listeners {
			for i := len(bus.listeners) - 1; i >= 0; i-- {
				if bus.listeners[i] == ch {
					// consume all data for this bus
				consume:
					for {
						select {
						case <-ch:
						default:
							break consume
						}
					}
					close(bus.listeners[i])
					bus.listeners = append(bus.listeners[:i], bus.listeners[i+1:]...)
					bus.wg.Done()
					break
				}
			}
		}
	}
	bus.mu.Unlock()
}

// Raise raises an event
func (bus *EventBus) Raise(events ...Event) {
	bus.mu.Lock()
	if !bus.closed {
		for i := 0; i < len(events); i++ {
			if events[i] == nil {
				continue
			}
			if bus.events != nil {
				bus.events.Value = events[i]
				bus.events = bus.events.Next()
			}
			for _, ch := range bus.listeners {
				ch <- events[i]
			}
		}
	}
	bus.mu.Unlock()
}

// Close stops new events from being trigerred, however the listeners need to close them self
func (bus *EventBus) Close() {
	bus.mu.Lock()
	if !bus.closed {
		bus.closed = true
	}
	bus.mu.Unlock()
}

// Wait waits until all listeners have been removed
func (bus *EventBus) Wait() {
	bus.wg.Wait()
}

// WaitAndClose waits until all listeners have been removed and closes the bus
func (bus *EventBus) WaitAndClose() {
	bus.Wait()
	bus.Close()
}

// Closed returns if the
func (bus *EventBus) Closed() bool {
	return bus.closed
}
