package query

type DataSource interface {
	Open() error
	Close() error
	Query(query string) (DataReader, error)
}

type DataReader interface {
	Next() bool
	Value() any
	Error() error
}

type DataWriter interface {
	Write(any) error
}
