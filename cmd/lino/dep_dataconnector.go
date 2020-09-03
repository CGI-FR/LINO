package main

import (
	infra "makeit.imfr.cgi.com/lino/internal/infra/dataconnector"
	domain "makeit.imfr.cgi.com/lino/pkg/dataconnector"
)

func dataconnectorStorage() domain.Storage {
	return infra.NewYAMLStorage()
}

func dataPingerFactory() map[string]domain.DataPingerFactory {
	return map[string]domain.DataPingerFactory{
		"postgres": infra.NewSQLDataPingerFactory(logger),
	}
}
