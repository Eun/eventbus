# eventbus [![Travis](https://img.shields.io/travis/Eun/eventbus.svg)](https://travis-ci.org/Eun/eventbus) [![Codecov](https://img.shields.io/codecov/c/github/Eun/eventbus.svg)](https://codecov.io/gh/Eun/eventbus) [![GoDoc](https://godoc.org/github.com/Eun/eventbus?status.svg)](https://godoc.org/github.com/Eun/eventbus) [![go-report](https://goreportcard.com/badge/github.com/Eun/eventbus)](https://goreportcard.com/report/github.com/Eun/eventbus)
A Multiconsumer/multiproducer bus.

```
go get gopkg.in/Eun/eventbus/v1
```

## Example
```go
package main

import (
	"fmt"

	"github.com/Eun/eventbus"
)

type IntegerEvent struct {
	I int
}

type ExitEvent struct{}

func consumer(consumerID int, bus *eventbus.EventBus, listener <-chan eventbus.Event) {
	defer bus.RemoveListener(listener)
	for ev := range listener {
		switch event := ev.(type) {
		case IntegerEvent:
			fmt.Printf("consumer %d: %d\n", consumerID, event.I)
		default:
			fmt.Printf("consumer %d: exit\n", consumerID)
			return
		}
	}
}

func main() {
	bus := eventbus.New()

	// add two consumers
	go consumer(1, bus, bus.AddListener())
	go consumer(2, bus, bus.AddListener())

	// raise 100 events
	for i := 0; i < 100; i++ {
		bus.Raise(IntegerEvent{I: i})
	}

	// add a third consumer, this one will only receive the last 10 items
	// because the HistorySize of the bus is set to 10
	go consumer(3, bus, bus.AddListener())

	// raise the exit event
	bus.Raise(ExitEvent{})

	// wait until all listeners have been removed
	bus.WaitAndClose()
}
```