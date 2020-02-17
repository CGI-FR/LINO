package main

import (
	"os"

	infra "makeit.imfr.cgi.com/lino/internal/infra/extract"
	domain "makeit.imfr.cgi.com/lino/pkg/extract"
)

func extractDataSourceFactory() map[string]domain.DataSourceFactory {
	return map[string]domain.DataSourceFactory{
		"postgres": infra.NewPostgresDataSourceFactory(logger),
		"godror": infra.NewOracleDataSourceFactory(logger),
	}
}

func extractRowExporter(file *os.File) domain.RowExporter {
	return infra.NewJSONRowExporter(file)
}

func traceListner(file *os.File) domain.TraceListener {
	return infra.NewJSONTraceListener(file)
}
