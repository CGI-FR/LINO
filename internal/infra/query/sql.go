package query

type DataReader struct{}

func (dr *DataReader) Next() bool {
	return false
}

func (dr *DataReader) Value() any {
	return nil
}

func (dr *DataReader) Error() error {
	return nil
}

type DataSource struct{}

func (ds *DataSource) Open() error {
	return nil
}

func (ds *DataSource) Close() error {
	return nil
}

func (ds *DataSource) Query(query string) (*DataReader, error) {
	return nil, nil
}
