package push

import (
	"encoding/csv"
	"fmt"
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
		// SQLLogger is not set.
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
	writer *csv.Writer
}

func (s *SQLLogger) OpenWriter(table push.Table, sqlquery string) *SQLLoggerWriter {
	if s == nil {
		// SQLLogger is not set.
		return nil
	}

	filename := s.folderPath + "/" + table.Name() + ".csv"
	writer, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot open file %v for SQL logger", filename)
	}

	logger := &SQLLoggerWriter{
		writer: csv.NewWriter(writer),
	}

	if _, err := writer.WriteString("# " + sqlquery + "\n"); err != nil {
		log.Warn().Err(err).Msgf("Cannot write into file %v for SQL logger", filename)
	}

	return logger
}

func (w *SQLLoggerWriter) Write(data []any) {
	if w == nil {
		// SQLLoggerWriter is not set.
		return
	}

	// Write the data to the file in CSV format
	if err := w.writer.Write(toStrings(data)); err != nil {
		log.Warn().Err(err).Msgf("Cannot log SQL statement")
	}
}

func toStrings(data []any) []string {
	strings := make([]string, len(data))
	for i, v := range data {
		strings[i] = toString(v)
	}
	return strings
}

func toString(data any) string {
	return fmt.Sprintf("%v", data)
}

func (w *SQLLoggerWriter) Close() {
	if w == nil {
		// SQLLoggerWriter is not set.
		return
	}

	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		log.Warn().Err(err).Msgf("Cannot flush SQL logger")
	}
}
