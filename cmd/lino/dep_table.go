package main

import (
	infra "makeit.imfr.cgi.com/lino/internal/infra/table"
	domain "makeit.imfr.cgi.com/lino/pkg/table"
)

func tableStorage() domain.Storage {
	return infra.NewYAMLStorage()
}

func tableExtractorFactory() map[string]domain.ExtractorFactory {
	return map[string]domain.ExtractorFactory{
		"postgres": infra.NewPostgresExtractorFactory(),
		"godror":   infra.NewOracleExtractorFactory(),
	}
}
