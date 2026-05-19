package query

type DataSource interface {
	Open() error
	Close() error
	Query(query string) (DataReader, error)
	SafeURL() string
}

type DataReader interface {
	Next() bool
	Value() any
	Error() error
}

type DataWriter interface {
	Write(any) error
}
