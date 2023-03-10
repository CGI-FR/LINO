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

package dataconnector

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/cgi-fr/lino/pkg/dataconnector"
)

// Version of the YAML strcuture
const Version string = "v1"

// YAMLStructure of the file
type YAMLStructure struct {
	Version        string              `yaml:"version"`
	DataConnectors []YAMLDataConnector `yaml:"dataconnectors,omitempty"`
}

// YAMLDataConnector defines how to store a dataconnector in YAML format
type YAMLDataConnector struct {
	Name     string           `yaml:"name"`
	URL      string           `yaml:"url"`
	ReadOnly bool             `yaml:"readonly"`
	Schema   string           `yaml:"schema,omitempty"`
	User     *YAMLValueHolder `yaml:"user,omitempty"`
	Password *YAMLValueHolder `yaml:"password,omitempty"`
}

type YAMLValueHolder struct {
	Value        *string `yaml:"value,omitempty"`
	ValueFromEnv *string `yaml:"valueFromEnv,omitempty"`
}

// NewYAMLStorage create a new YAML storage
func NewYAMLStorage() *YAMLStorage {
	return &YAMLStorage{}
}

// YAMLStorage provides storage in a local YAML file
type YAMLStorage struct{}

// List all dataconnector stored in the YAML file
func (s YAMLStorage) List() ([]dataconnector.DataConnector, *dataconnector.Error) {
	list, err := readFile()
	if err != nil {
		return nil, err
	}

	result := []dataconnector.DataConnector{}

	for _, ym := range list.DataConnectors {
		m := dataconnector.DataConnector{
			Name:     ym.Name,
			URL:      ym.URL,
			ReadOnly: ym.ReadOnly,
			Schema:   ym.Schema,
		}
		if ym.User != nil {
			if ym.User.Value != nil {
				m.User.Value = *ym.User.Value
			}
			if ym.User.ValueFromEnv != nil {
				m.User.ValueFromEnv = *ym.User.ValueFromEnv
			}
		}
		if ym.Password != nil {
			if ym.Password.ValueFromEnv != nil {
				m.Password.ValueFromEnv = *ym.Password.ValueFromEnv
			}
		}
		result = append(result, m)
	}

	return result, nil
}

// Store a dataconnector in the YAML file
func (s YAMLStorage) Store(m *dataconnector.DataConnector) *dataconnector.Error {
	list, err := readFile()
	if err != nil {
		return err
	}

	yml := YAMLDataConnector{
		Name:     m.Name,
		URL:      m.URL,
		ReadOnly: m.ReadOnly,
		Schema:   m.Schema,
		User:     nil,
		Password: nil,
	}

	if m.User.ValueFromEnv != "" || m.User.Value != "" {
		yml.User = &YAMLValueHolder{}
		if m.User.ValueFromEnv != "" {
			yml.User.ValueFromEnv = &m.User.ValueFromEnv
		}
		if m.User.Value != "" {
			yml.User.Value = &m.User.Value
		}
	}

	if m.Password.ValueFromEnv != "" {
		yml.Password = &YAMLValueHolder{}
		if m.Password.ValueFromEnv != "" {
			yml.Password.ValueFromEnv = &m.Password.ValueFromEnv
		}
	}

	list.DataConnectors = append(list.DataConnectors, yml)

	err = writeFile(list)
	if err != nil {
		return err
	}

	return nil
}

func readFile() (*YAMLStructure, *dataconnector.Error) {
	list := &YAMLStructure{
		Version: Version,
	}

	if _, err := os.Stat("dataconnector.yaml"); os.IsNotExist(err) {
		return list, nil
	}

	dat, err := os.ReadFile("dataconnector.yaml")
	if err != nil {
		return nil, &dataconnector.Error{Description: err.Error()}
	}

	err = yaml.Unmarshal(dat, list)
	if err != nil {
		return nil, &dataconnector.Error{Description: err.Error()}
	}

	if list.Version != Version {
		return nil, &dataconnector.Error{Description: "invalid version in ./dataconnector.yaml (" + list.Version + ")"}
	}

	return list, nil
}

func writeFile(list *YAMLStructure) *dataconnector.Error {
	out := &bytes.Buffer{}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)

	err := enc.Encode(list)
	if err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}

	err = os.WriteFile("dataconnector.yaml", out.Bytes(), 0o600)
	if err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}

	return nil
}
