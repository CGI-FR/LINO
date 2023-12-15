package analyse

import (
	"bufio"
	"io"

	"github.com/cgi-fr/lino/pkg/analyse"
	"github.com/cgi-fr/rimo/pkg/model"
	"gopkg.in/yaml.v3"
)

type StdWriter struct {
	out bufio.Writer
}

func NewStdWriter(out io.Writer) analyse.Writer {
	return StdWriter{
		out: *bufio.NewWriter(out),
	}
}

func (w StdWriter) Write(report *model.Base) error {
	bytes, err := yaml.Marshal(report)
	if err != nil {
		return err
	}

	_, err = w.out.Write(bytes)
	if err != nil {
		return err
	}

	err = w.out.Flush()
	if err != nil {
		return err
	}

	return nil
}
