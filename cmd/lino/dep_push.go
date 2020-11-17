package main

import (
	"io"

	infra "makeit.imfr.cgi.com/lino/internal/infra/push"
	domain "makeit.imfr.cgi.com/lino/pkg/push"
)

func pushDataDestinationFactory() map[string]domain.DataDestinationFactory {
	return map[string]domain.DataDestinationFactory{
		"postgres":   infra.NewPostgresDataDestinationFactory(logger),
		"godror":     infra.NewOracleDataDestinationFactory(logger),
		"godror-raw": infra.NewOracleDataDestinationFactory(logger),
	}
}

func pushRowIteratorFactory() func(io.ReadCloser) domain.RowIterator {
	return infra.NewJSONRowIterator
}

func pushRowExporterFactory() func(io.Writer) domain.RowWriter {
	return infra.NewJSONRowWriter
}
