package eventbus_test

import (
	"testing"

	"github.com/Eun/eventbus"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

type testEvent struct {
	i int
}

func (e *testEvent) Source() string {
	return ""
}

func failIfGotEvent(t *testing.T, eventChan <-chan eventbus.Event) {
	select {
	case ev, ok := <-eventChan:
		if ok {
			require.FailNow(t, "eventChan should not have any values", "was %s", spew.Sdump(ev))
		}
	default:
	}
}

func failIfGotNoEvent(t *testing.T, eventChan <-chan eventbus.Event) {
	select {
	case _, ok := <-eventChan:
		if !ok {
			require.FailNow(t, "eventChan should have a value")
		}
	default:
		require.FailNow(t, "eventChan should have a value")
	}
}

func TestHistory(t *testing.T) {
	historySize := 6
	bus := eventbus.New(historySize)
	require.Equal(t, historySize, bus.HistorySize())

	for i := 0; i < historySize*2; i++ {
		bus.Raise(&testEvent{i})
	}

	eventChan := bus.AddListener()

	for i := historySize; i < historySize*2; i++ {
		ev := <-eventChan
		require.Equal(t, i, ev.(*testEvent).i)
	}
	bus.Close()
}

func TestListener(t *testing.T) {
	channelSize := 10
	bus := eventbus.New(0)
	require.Equal(t, 0, bus.HistorySize())
	eventChan := bus.AddListener(channelSize)

	for i := 0; i < channelSize; i++ {
		bus.Raise(&testEvent{i})
	}

	for i := 0; i < channelSize; i++ {
		ev := <-eventChan
		require.Equal(t, i, ev.(*testEvent).i)
	}

	failIfGotEvent(t, eventChan)

	bus.RemoveListener(eventChan)

	failIfGotEvent(t, eventChan)
	bus.Close()
}

func TestRemoveListenerWithOutstandingData(t *testing.T) {
	channelSize := 10
	bus := eventbus.New(0)
	require.Equal(t, 0, bus.HistorySize())
	eventChan := bus.AddListener(channelSize)

	for i := 0; i < channelSize; i++ {
		bus.Raise(&testEvent{i})
	}

	for i := 0; i < channelSize/2; i++ {
		ev := <-eventChan
		require.Equal(t, i, ev.(*testEvent).i)
	}

	bus.RemoveListener(eventChan)

	failIfGotEvent(t, eventChan)
	bus.Close()
}

func TestAddListenerOnClosedBus(t *testing.T) {
	bus := eventbus.New()
	bus.Close()
	require.Nil(t, bus.AddListener())
}
func TestRaiseOnClosedBus(t *testing.T) {
	bus := eventbus.New()
	eventChan := bus.AddListener()
	bus.Close()
	bus.Raise(&testEvent{0})
	failIfGotEvent(t, eventChan)
}
func TestRaiseNilEvent(t *testing.T) {
	bus := eventbus.New()
	eventChan := bus.AddListener()
	bus.Raise(nil)
	failIfGotEvent(t, eventChan)
	bus.Close()
}

func TestCloseOnClosedBus(t *testing.T) {
	bus := eventbus.New()
	bus.Close()
	require.True(t, bus.Closed())
	bus.Close()
}

func TestWaitAndClose(t *testing.T) {
	bus := eventbus.New()

	eventChan := bus.AddListener()
	go func() {
		failIfGotNoEvent(t, eventChan)
		bus.RemoveListener(eventChan)
	}()
	bus.Raise(&testEvent{0})
	bus.WaitAndClose()
}
