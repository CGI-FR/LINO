package id

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
	"makeit.imfr.cgi.com/lino/pkg/id"
)

// Version of the YAML strcuture.
const Version string = "v1"

// YAMLStructure of the file.
type YAMLStructure struct {
	Version           string                `yaml:"version"`
	IngressDescriptor YAMLIngressDescriptor `yaml:"IngressDescriptor"`
}

// YAMLIngressDescriptor defines how to store an ingress descriptor in YAML format.
type YAMLIngressDescriptor struct {
	StartTable string         `yaml:"startTable"`
	Relations  []YAMLRelation `yaml:"relations"`
}

// YAMLRelation defines how to store a relation in YAML format.
type YAMLRelation struct {
	Name   string    `yaml:"name"`
	Parent YAMLTable `yaml:"parent"`
	Child  YAMLTable `yaml:"child"`
}

// YAMLTable defines how to store a table in YAML format.
type YAMLTable struct {
	Name   string `yaml:"name"`
	Lookup bool   `yaml:"lookup"`
}

// YAMLStorage provides storage in a local YAML file
type YAMLStorage struct {
}

// NewYAMLStorage create a new YAML storage
func NewYAMLStorage() *YAMLStorage {
	return &YAMLStorage{}
}

// Store ingress descriptor in the YAML file
func (s *YAMLStorage) Store(id id.IngressDescriptor) *id.Error {
	structure := YAMLStructure{
		Version: Version,
	}

	relations := []YAMLRelation{}
	list := id.Relations()
	for i := uint(0); i < list.Len(); i++ {
		relation := list.Relation(i)
		relations = append(relations, YAMLRelation{
			Name:   relation.Name(),
			Parent: YAMLTable{Name: relation.Parent().Name(), Lookup: relation.LookUpParent()},
			Child:  YAMLTable{Name: relation.Child().Name(), Lookup: relation.LookUpChild()},
		})
	}

	structure.IngressDescriptor = YAMLIngressDescriptor{
		StartTable: id.StartTable().Name(),
		Relations:  relations,
	}

	err := writeFile(&structure)
	if err != nil {
		return err
	}

	return nil
}

func (s *YAMLStorage) Read() (id.IngressDescriptor, *id.Error) {
	structure, err := readFile()
	if err != nil {
		return nil, err
	}

	relations := []id.IngressRelation{}
	for _, relation := range structure.IngressDescriptor.Relations {
		relations = append(relations,
			id.NewIngressRelation(
				id.NewRelation(
					relation.Name,
					id.NewTable(relation.Parent.Name),
					id.NewTable(relation.Child.Name),
				),
				relation.Parent.Lookup, relation.Child.Lookup),
		)
	}

	return id.NewIngressDescriptor(id.NewTable(structure.IngressDescriptor.StartTable), id.NewIngressRelationList(relations)), nil
}

func writeFile(structure *YAMLStructure) *id.Error {
	out, err := yaml.Marshal(structure)
	if err != nil {
		return &id.Error{Description: err.Error()}
	}

	err = ioutil.WriteFile("ingress-descriptor.yaml", out, 0640)
	if err != nil {
		return &id.Error{Description: err.Error()}
	}

	return nil
}

func readFile() (*YAMLStructure, *id.Error) {
	structure := &YAMLStructure{
		Version: Version,
	}

	dat, err := ioutil.ReadFile("ingress-descriptor.yaml")
	if err != nil {
		return nil, &id.Error{Description: err.Error()}
	}

	err = yaml.Unmarshal(dat, structure)
	if err != nil {
		return nil, &id.Error{Description: err.Error()}
	}

	if structure.Version != Version {
		return nil, &id.Error{Description: "invalid version in ./ingress-descriptor.yaml (" + structure.Version + ")"}
	}

	return structure, nil
}
