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

package sequence

import (
	"bytes"
	"io/ioutil"

	"github.com/cgi-fr/lino/pkg/sequence"
	"gopkg.in/yaml.v3"
)

// Version of the YAML strcuture.
const Version string = "v1"

// YAMLStructure of the file.
type YAMLStructure struct {
	Version  string         `yaml:"version"`
	Sequence []YAMLSequence `yaml:"sequences,omitempty"`
}

// YAMLSequence defines how to store a sequence
type YAMLSequence struct {
	Name   string `yaml:"name"`
	Table  string `yaml:"table"`
	Column string `yaml:"column"`
}

// YAMLStorage provides storage in a local YAML file
type YAMLStorage struct{}

// NewYAMLStorage create a new YAML storage
func NewYAMLStorage() *YAMLStorage {
	return &YAMLStorage{}
}

// List all tables stored in the YAML file
func (s YAMLStorage) List() ([]sequence.Sequence, *sequence.Error) {
	list, err := readFile()
	if err != nil {
		return nil, err
	}
	result := []sequence.Sequence{}

	for _, ym := range list.Sequence {
		s := sequence.Sequence{
			Name:   ym.Name,
			Table:  ym.Table,
			Column: ym.Column,
		}

		result = append(result, s)
	}

	return result, nil
}

// Store tables in the YAML file
func (s YAMLStorage) Store(sequences []sequence.Sequence) *sequence.Error {
	list := YAMLStructure{
		Version: Version,
	}

	for _, r := range sequences {
		yml := YAMLSequence{
			Name:   r.Name,
			Table:  r.Table,
			Column: r.Column,
		}
		list.Sequence = append(list.Sequence, yml)
	}

	if err := writeFile(&list); err != nil {
		return err
	}

	return nil
}

func readFile() (*YAMLStructure, *sequence.Error) {
	list := &YAMLStructure{
		Version: Version,
	}

	dat, err := ioutil.ReadFile("sequences.yaml")
	if err != nil {
		return nil, &sequence.Error{Description: err.Error()}
	}

	err = yaml.Unmarshal(dat, list)
	if err != nil {
		return nil, &sequence.Error{Description: err.Error()}
	}

	if list.Version != Version {
		return nil, &sequence.Error{Description: "invalid version in ./sequences.yaml (" + list.Version + ")"}
	}

	return list, nil
}

func writeFile(list *YAMLStructure) *sequence.Error {
	out := &bytes.Buffer{}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)

	err := enc.Encode(list)
	if err != nil {
		return &sequence.Error{Description: err.Error()}
	}

	err = ioutil.WriteFile("sequences.yaml", out.Bytes(), 0600)
	if err != nil {
		return &sequence.Error{Description: err.Error()}
	}
	return nil
}
