package main

import (
	"os"

	infra "makeit.imfr.cgi.com/lino/internal/infra/push"
	domain "makeit.imfr.cgi.com/lino/pkg/push"
)

func pushDataDestinationFactory() map[string]domain.DataDestinationFactory {
	return map[string]domain.DataDestinationFactory{
		"postgres": infra.NewPostgresDataDestinationFactory(logger),
		"godror":   infra.NewOracleDataDestinationFactory(logger),
	}
}

func pushRowIterator(file *os.File) domain.RowIterator {
	return infra.NewJSONRowIterator(file)
}
