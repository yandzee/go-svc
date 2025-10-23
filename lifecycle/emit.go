package lifecycle

type EventKind string

const (
	RunInit         EventKind = "run-init"
	RunInitError    EventKind = "run-init-error"
	Running         EventKind = "running"
	RunEvent        EventKind = "run-event"
	RunFinished     EventKind = "run-finished"
	ShutdownRequest EventKind = "shutdown-request"
)

type Emit[T any] struct {
	enabled bool
	Channel chan Event[T]
}

type Event[T any] struct {
	Kind   EventKind `json:"kind"`
	Error  error     `json:"error,omitempty"`
	Detail T         `json:"detail,omitempty"`
}

func New[T any](enabled bool) *Emit[T] {
	s := &Emit[T]{
		enabled: enabled,
	}

	if enabled {
		s.Channel = make(chan Event[T], 1)
	}

	return s
}

func (s *Emit[T]) RunInit() {
	s.emitIfEnabled(RunInit, nil)
}

func (s *Emit[T]) RunInitError(err error) {
	s.emitIfEnabled(RunInitError, err)
}

func (s *Emit[T]) Running() {
	s.emitIfEnabled(Running, nil)
}

func (s *Emit[T]) RunLoopEvent(d T) {
	s.emitIfEnabled(RunEvent, nil, d)
}

func (s *Emit[T]) RunFinished(err error) {
	s.emitIfEnabled(RunFinished, err)
}

func (s *Emit[T]) ShutdownRequest() {
	s.emitIfEnabled(ShutdownRequest, nil)
}

func (s *Emit[T]) emitIfEnabled(kind EventKind, err error, d ...T) {
	if !s.enabled {
		return
	}

	evt := Event[T]{
		Kind:  kind,
		Error: err,
	}

	if len(d) > 0 {
		evt.Detail = d[0]
	}

	s.Channel <- evt
}
