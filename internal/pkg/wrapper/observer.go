package wrapper

type InstanceEventType uint

const (
	EventExit InstanceEventType = iota
	EventStart
)

// Event defines an indication of a some occurence
type InstanceEvent struct {
	Origin     *TTYVNTInstance
	Type       InstanceEventType
	ExitReason error
}

func NewExitInstanceEvent(origin *TTYVNTInstance, reason error) InstanceEvent {
	return InstanceEvent{
		Origin:     origin,
		Type:       EventExit,
		ExitReason: reason,
	}
}

func NewStartInstanceEvent(origin *TTYVNTInstance) InstanceEvent {
	return InstanceEvent{
		Origin:     origin,
		Type:       EventStart,
		ExitReason: nil,
	}
}

// Observer defines a standard interface
// to listen for a specific event.
type InstanceObserver interface {
	// OnNotify allows to publsh an event
	OnNotify(InstanceEvent)
}

// Notifier is the instance being observed.
type InstanceNotifier interface {
	// Register itself to listen/observe events.
	Register(InstanceObserver)
	// Remove itself from the collection of observers/listeners.
	Unregister(InstanceObserver)
	// Notify publishes new events to listeners.
	Notify(InstanceEvent)
}
