package main

import (
	infra "makeit.imfr.cgi.com/lino/internal/infra/dataconnector"
	domain "makeit.imfr.cgi.com/lino/pkg/dataconnector"
)

func dataconnectorStorage() domain.Storage {
	return infra.NewYAMLStorage()
}
