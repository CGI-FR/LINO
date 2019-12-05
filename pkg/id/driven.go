package id

// Storage allows to store and retrieve ingress descriptor objects.
type Storage interface {
	Store(IngressDescriptor) *Error
	Read() (IngressDescriptor, *Error)
}

// RelationReader read relations from a source.
type RelationReader interface {
	Read() (RelationList, *Error)
}

// Exporter export the extraction plan.
type Exporter interface {
	Export(ExtractionPlan) *Error
}

// Logger for events.
type Logger interface {
	Trace(msg string)
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// Nologger default implementation do nothing.
type Nologger struct{}

// Trace event.
func (l Nologger) Trace(msg string) {}

// Debug event.
func (l Nologger) Debug(msg string) {}

// Info event.
func (l Nologger) Info(msg string) {}

// Warn event.
func (l Nologger) Warn(msg string) {}

// Error event.
func (l Nologger) Error(msg string) {}
