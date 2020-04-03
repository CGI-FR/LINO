package pull

// RowExporter receives pulled rows one by one.
type RowExporter interface {
	Export(Row) *Error
}

// DataSourceFactory exposes methods to create new datasources.
type DataSourceFactory interface {
	New(url string) DataSource
}

// DataSource to read in the pull process.
type DataSource interface {
	Open() *Error
	RowReader(source Table, filter Filter) (RowReader, *Error)
	Close() *Error
}

// RowReader over DataSource.
type RowReader interface {
	Next() bool
	Value() (Row, *Error)
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

// TraceListener receives diagnostic trace
type TraceListener interface {
	TraceStep(Step, Filter) TraceListener
}

// NoTraceListener default implementation do nothing.
type NoTraceListener struct{}

// TraceStep catch Step event.
func (t NoTraceListener) TraceStep(s Step, filter Filter) TraceListener { return t }
