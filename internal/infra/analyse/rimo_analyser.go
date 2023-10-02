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

package analyse

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/cgi-fr/lino/pkg/analyse"
	"github.com/cgi-fr/rimo/pkg/model"
	"github.com/cgi-fr/rimo/pkg/rimo"
)

type RimoAnalyserFactory struct{}

func (r RimoAnalyserFactory) New(out io.Writer) analyse.Analyser {
	return RimoAnalyser{
		writer: &YAMLWriter{output: out},
	}
}

type RimoAnalyser struct {
	writer rimo.Writer
}

func (ra RimoAnalyser) Analyse(ds *analyse.ColumnIterator) error {
	return rimo.AnalyseBase(ds, ra)
}

func (ra RimoAnalyser) Export(base *model.Base) error {
	return ra.writer.Export(base)
}

// YAML Writter interface

type YAMLWriter struct {
	output io.Writer
}

// Write a YAML file from RIMO base at outputPath.
func (w *YAMLWriter) Export(base *model.Base) error {
	// Encode Base to YAML.
	encoder := yaml.NewEncoder(w.output)
	defer encoder.Close()

	err := encoder.Encode(base)
	if err != nil {
		return fmt.Errorf("failed to encode Base to YAML: %w", err)
	}

	return nil
}
