package main

import (
	"io"
	"os"

	infra "makeit.imfr.cgi.com/lino/internal/infra/pull"
	domain "makeit.imfr.cgi.com/lino/pkg/pull"
)

func pullDataSourceFactory() map[string]domain.DataSourceFactory {
	return map[string]domain.DataSourceFactory{
		"postgres": infra.NewPostgresDataSourceFactory(logger),
	}
}

func pullRowExporterFactory() func(file io.Writer) domain.RowExporter {
	return func(file io.Writer) domain.RowExporter {
		return infra.NewJSONRowExporter(file)
	}
}

func pullRowReaderFactory() func(file io.ReadCloser) domain.RowReader {
	return func(file io.ReadCloser) domain.RowReader {
		return infra.NewJSONRowReader(file)
	}
}

func traceListner(file *os.File) domain.TraceListener {
	return infra.NewJSONTraceListener(file)
}
