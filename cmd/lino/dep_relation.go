package main

import (
	infra "makeit.imfr.cgi.com/lino/internal/infra/relation"
	domain "makeit.imfr.cgi.com/lino/pkg/relation"
)

func relationStorage() domain.Storage {
	return infra.NewYAMLStorage()
}

func relationExtractorFactory() map[string]domain.ExtractorFactory {
	return map[string]domain.ExtractorFactory{
		"postgres":   infra.NewPostgresExtractorFactory(),
		"godror":     infra.NewOracleExtractorFactory(),
		"godror-raw": infra.NewOracleExtractorFactory(),
	}
}
