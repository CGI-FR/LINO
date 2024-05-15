package query

import "fmt"

type Driver struct {
	datasource DataSource
	writer     DataWriter
}

func NewDriver(datasource DataSource, writer DataWriter) *Driver {
	return &Driver{datasource, writer}
}

func (d *Driver) Open() error {
	if err := d.datasource.Open(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (d *Driver) Close() error {
	if err := d.datasource.Close(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (d *Driver) Execute(query string) error {
	reader, err := d.datasource.Query(query)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if reader == nil {
		return nil
	}

	for reader.Next() {
		if err := d.writer.Write(reader.Value()); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if err = reader.Error(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
