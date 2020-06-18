package table

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
	"makeit.imfr.cgi.com/lino/pkg/table"
)

// Version of the YAML strcuture.
const Version string = "v1"

// YAMLStructure of the file.
type YAMLStructure struct {
	Version string      `yaml:"version"`
	Tables  []YAMLTable `yaml:"tables,omitempty"`
}

// YAMLTable defines how to store a table in YAML format.
type YAMLTable struct {
	Name string   `yaml:"name"`
	Keys []string `yaml:"keys"`
}

// YAMLStorage provides storage in a local YAML file
type YAMLStorage struct{}

// NewYAMLStorage create a new YAML storage
func NewYAMLStorage() *YAMLStorage {
	return &YAMLStorage{}
}

// List all tables stored in the YAML file
func (s YAMLStorage) List() ([]table.Table, *table.Error) {
	list, err := readFile()
	if err != nil {
		return nil, err
	}
	result := []table.Table{}

	for _, ym := range list.Tables {
		m := table.Table{
			Name: ym.Name,
			Keys: ym.Keys,
		}
		result = append(result, m)
	}

	return result, nil
}

// Store tables in the YAML file
func (s YAMLStorage) Store(tables []table.Table) *table.Error {
	list := YAMLStructure{
		Version: Version,
	}

	for _, r := range tables {
		yml := YAMLTable{
			Name: r.Name,
			Keys: r.Keys,
		}
		list.Tables = append(list.Tables, yml)
	}

	err := writeFile(&list)
	if err != nil {
		return err
	}

	return nil
}

func readFile() (*YAMLStructure, *table.Error) {
	list := &YAMLStructure{
		Version: Version,
	}

	dat, err := ioutil.ReadFile("tables.yaml")
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	err = yaml.Unmarshal(dat, list)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	if list.Version != Version {
		return nil, &table.Error{Description: "invalid version in ./tables.yaml (" + list.Version + ")"}
	}

	return list, nil
}

func writeFile(list *YAMLStructure) *table.Error {
	out, err := yaml.Marshal(list)
	if err != nil {
		return &table.Error{Description: err.Error()}
	}

	err = ioutil.WriteFile("tables.yaml", out, 0600)
	if err != nil {
		return &table.Error{Description: err.Error()}
	}

	return nil
}
