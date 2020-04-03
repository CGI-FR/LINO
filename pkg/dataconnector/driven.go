package dataconnector

// Storage allows to store and retrieve DataConnector objects.
type Storage interface {
	List() ([]DataConnector, *Error)
	Store(*DataConnector) *Error
}

// DataPingerFactory create a DataPing for the given `url`
type DataPingerFactory interface {
	New(url string) DataPinger
}

// Datapinger test connection
type DataPinger interface {
	Ping() *Error
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
