package main

import (
	"io"

	infra "makeit.imfr.cgi.com/lino/internal/infra/push"
	domain "makeit.imfr.cgi.com/lino/pkg/push"
)

func pushDataDestinationFactory() map[string]domain.DataDestinationFactory {
	return map[string]domain.DataDestinationFactory{
		"postgres": infra.NewPostgresDataDestinationFactory(logger),
	}
}

func pushRowIteratorFactory() func(io.ReadCloser) domain.RowIterator {
	return infra.NewJSONRowIterator
}
