package main

import (
	"os"

	infra "makeit.imfr.cgi.com/lino/internal/infra/id"
	domain "makeit.imfr.cgi.com/lino/pkg/id"
)

func idStorage() domain.Storage {
	return infra.NewMultiStorage(infra.NewYAMLStorage(), infra.NewDOTStorage())
}

func idExporter() domain.Exporter {
	return infra.NewGraphVizExporter()
}

func idJSONStorage(file os.File) domain.Storage {
	return infra.NewJSONStorage(file)
}
