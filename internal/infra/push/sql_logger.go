package push

import (
	"io"
	"os"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/rs/zerolog/log"
)

type SQLLogger struct {
	folderPath string
}

func NewSQLLogger(folderPath string) *SQLLogger {
	return &SQLLogger{
		folderPath: folderPath,
	}
}

func (s *SQLLogger) Open() error {
	if s == nil {
		// SQLLogger is not set, it is ok to return no error
		// because the caller will not use it.
		return nil
	}

	// Check if the folder exists
	if _, err := os.Stat(s.folderPath); os.IsNotExist(err) {
		// Create the folder if it doesn't exist
		if err := os.MkdirAll(s.folderPath, 0o755); err != nil { //nolint:mnd
			return err
		}
	}
	// Check if we have permission to write in the folder
	if err := os.WriteFile(s.folderPath+"/.test", []byte("test"), 0o600); err != nil {
		return err
	}
	// Remove the test file
	_ = os.Remove(s.folderPath + "/.test")
	return nil
}

type SQLLoggerWriter struct {
	table  push.Table
	writer io.Writer
}

func (s *SQLLogger) OpenWriter(table push.Table, sqlquery string) *SQLLoggerWriter {
	filename := s.folderPath + "/" + table.Name() + ".csv"
	writer, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot open file %v for SQL logger", filename)
	}

	logger := &SQLLoggerWriter{
		table:  table,
		writer: writer, //nolint:mnd
	}

	writer.WriteString("# " + sqlquery + "\n")

	return logger
}
