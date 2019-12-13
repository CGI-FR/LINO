package main

import (
	"os"

	infra "makeit.imfr.cgi.com/lino/internal/infra/load"
	domain "makeit.imfr.cgi.com/lino/pkg/load"
)

func loadDataDestinationFactory() map[string]domain.DataDestinationFactory {
	return map[string]domain.DataDestinationFactory{
		"postgres": infra.NewPostgresDataDestinationFactory(logger),
	}
}

func loadRowIterator(file *os.File) domain.RowIterator {
	return infra.NewJSONRowIterator(file)
}
