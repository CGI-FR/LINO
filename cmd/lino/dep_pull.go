// Copyright (C) 2021 CGI France
//
// This file is part of LINO.
//
// LINO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LINO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LINO.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"io"
	"os"
	"time"

	infra "github.com/cgi-fr/lino/internal/infra/pull"
	domain "github.com/cgi-fr/lino/pkg/pull"
)

func pullDataSourceFactory() map[string]domain.DataSourceFactory {
	return map[string]domain.DataSourceFactory{
		"postgres":   infra.NewPostgresDataSourceFactory(),
		"godror":     infra.NewOracleDataSourceFactory(),
		"godror-raw": infra.NewOracleDataSourceFactory(),
		"mysql":      infra.NewMariadbDataSourceFactory(),
		"db2":        infra.NewDb2DataSourceFactory(),
		"http":       infra.NewHTTPDataSourceFactory(),
		"ws":         infra.NewWSDataSourceFactory(),
		"sqlserver":  infra.NewSQLServerDataSourceFactory(),
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

func pullKeyStoreFactory() func(file io.ReadCloser, keys []string) (domain.KeyStore, error) {
	return func(file io.ReadCloser, keys []string) (domain.KeyStore, error) {
		return infra.NewJSONKeyStore(file, keys)
	}
}

func traceListner(file *os.File) domain.TraceListener {
	return infra.NewJSONTraceListener(file)
}

func pullMaxLifeTime(maxLifetimeInSeconds int64) domain.DataSourceOption {
	return infra.WithMaxLifetime(time.Duration(maxLifetimeInSeconds) * time.Second)
}

func pullMaxOpenConns(maxOpenConns int) domain.DataSourceOption {
	return infra.WithMaxOpenConns(maxOpenConns)
}

func pullMaxIdleConns(maxIdleConns int) domain.DataSourceOption {
	return infra.WithMaxIdleConns(maxIdleConns)
}
