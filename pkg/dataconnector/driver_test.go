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

package dataconnector_test

import (
	"fmt"
	"testing"

	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/stretchr/testify/assert"
)

// MemoryStorage provides storage of DataConnector in memory
type MemoryStorage struct {
	repo []dataconnector.DataConnector
}

// List all dataconnector stored in memory
func (s *MemoryStorage) List() ([]dataconnector.DataConnector, *dataconnector.Error) {
	return s.repo, nil
}

// Store a dataconnector in memory
func (s *MemoryStorage) Store(m *dataconnector.DataConnector) *dataconnector.Error {
	fmt.Println("Appending", m)
	s.repo = append(s.repo, *m)
	return nil
}

// ErrorStorage always return an error
type ErrorStorage struct {
	ListError  *dataconnector.Error
	StoreError *dataconnector.Error
}

// List all dataconnector stored in memory
func (s *ErrorStorage) List() ([]dataconnector.DataConnector, *dataconnector.Error) {
	return nil, s.ListError
}

// Store a dataconnector in memory
func (s *ErrorStorage) Store(m *dataconnector.DataConnector) *dataconnector.Error {
	return s.StoreError
}

func TestAddToNonEmptyStorage(t *testing.T) {
	storage := &MemoryStorage{repo: []dataconnector.DataConnector{dataconnector.DataConnector{Name: "First", URL: "test://localhost:1234"}}}
	alias := &dataconnector.DataConnector{Name: "Second", URL: "test://localhost:1234"}

	err := dataconnector.Add(storage, alias)
	aliases, _ := storage.List()

	assert.Nil(t, err, "An error occurred while using Add method")
	assert.Equal(t, 2, len(aliases), "Two alias should be stored")
}

func TestAddExistingToStorage(t *testing.T) {
	storage := &MemoryStorage{repo: []dataconnector.DataConnector{dataconnector.DataConnector{Name: "Exists", URL: "test://localhost:1234"}}}
	alias := &dataconnector.DataConnector{Name: "Exists", URL: "test://localhost:1234"}

	err := dataconnector.Add(storage, alias)
	aliases, _ := storage.List()

	assert.Nil(t, err, "An error occurred while using Add method")
	assert.Equal(t, 1, len(aliases), "Only one alias should be stored")
}

func TestAddStorageListErrorHandling(t *testing.T) {
	storage := &ErrorStorage{&dataconnector.Error{Description: "ListError"}, nil}
	alias := &dataconnector.DataConnector{Name: "Test", URL: "test://localhost:1234"}

	err := dataconnector.Add(storage, alias)

	assert.NotNil(t, err, "Add method should return an error")
	assert.Equal(t, "ListError", err.Error(), "Error should contain the description provided by storage List method")
}

func TestAddStorageStoreErrorHandling(t *testing.T) {
	storage := &ErrorStorage{nil, &dataconnector.Error{Description: "StoreError"}}
	alias := &dataconnector.DataConnector{Name: "Test", URL: "test://localhost:1234"}

	err := dataconnector.Add(storage, alias)

	assert.NotNil(t, err, "Add method should return an error")
	assert.Equal(t, "StoreError", err.Error(), "Error should contain the description provided by storage Store method")
}

func TestGetFromEmptyStorage(t *testing.T) {
	storage := &MemoryStorage{}
	name := "Test"

	alias, err := dataconnector.Get(storage, name)

	assert.Nil(t, err, "An error occurred while using Get method")
	assert.Nil(t, alias, "No alias should be returned")
}

func TestGetFromNonEmptyStorage(t *testing.T) {
	storage := &MemoryStorage{repo: []dataconnector.DataConnector{dataconnector.DataConnector{Name: "First", URL: "test://localhost:1234"}, dataconnector.DataConnector{Name: "Second", URL: "test://localhost:5678"}}}
	name := "First"

	alias, err := dataconnector.Get(storage, name)

	assert.Nil(t, err, "An error occurred while using Get method")
	assert.NotNil(t, alias, "An alias should be returned")
	assert.Equal(t, name, alias.Name, "Name of returned alias is invalid")
	assert.Equal(t, "test://localhost:1234", alias.URL, "URL of returned alias is invalid")
}

func TestGetNonExistingFromNonEmptyStorage(t *testing.T) {
	storage := &MemoryStorage{repo: []dataconnector.DataConnector{dataconnector.DataConnector{Name: "First", URL: "test://localhost:1234"}, dataconnector.DataConnector{Name: "Second", URL: "test://localhost:5678"}}}
	name := "Third"

	alias, err := dataconnector.Get(storage, name)

	assert.Nil(t, err, "An error occurred while using Get method")
	assert.Nil(t, alias, "No alias should be returned")
}

func TestGetStorageListErrorHandling(t *testing.T) {
	storage := &ErrorStorage{&dataconnector.Error{Description: "ListError"}, nil}
	name := "Test"

	alias, err := dataconnector.Get(storage, name)

	assert.NotNil(t, err, "Get method should return an error")
	assert.Nil(t, alias, "No alias should be returned")
	assert.Equal(t, "ListError", err.Error(), "Error should contain the description provided by storage Store method")
}

func TestListFromEmptyStorage(t *testing.T) {
	storage := &MemoryStorage{}

	aliases, err := dataconnector.List(storage)

	assert.Nil(t, err, "An error occurred while using List method")
	assert.Empty(t, aliases, "An empty alias list should be returned")
}

func TestListFromNonEmptyStorage(t *testing.T) {
	storage := &MemoryStorage{repo: []dataconnector.DataConnector{dataconnector.DataConnector{Name: "First", URL: "test://localhost:1234"}, dataconnector.DataConnector{Name: "Second", URL: "test://localhost:5678"}}}

	aliases, err := dataconnector.List(storage)

	assert.Nil(t, err, "An error occurred while using List method")
	assert.Len(t, aliases, 2, "The aliases list should be returned")
}

func TestListStorageListErrorHandling(t *testing.T) {
	storage := &ErrorStorage{&dataconnector.Error{Description: "ListError"}, nil}

	aliases, err := dataconnector.List(storage)

	assert.NotNil(t, err, "List method should return an error")
	assert.Nil(t, aliases, "No alias list should be returned")
	assert.Equal(t, "ListError", err.Error(), "Error should contain the description provided by storage Store method")
}

func TestAdd(t *testing.T) {
	type args struct {
		s dataconnector.Storage
		m *dataconnector.DataConnector
	}
	tests := []struct {
		name string
		args args
		want *dataconnector.Error
	}{
		struct {
			name string
			args args
			want *dataconnector.Error
		}{name: "Empty storage", args: args{
			&MemoryStorage{},
			&dataconnector.DataConnector{Name: "Test", URL: "test://localhost:1234"},
		}, want: nil},
		struct {
			name string
			args args
			want *dataconnector.Error
		}{name: "read only alias", args: args{
			&MemoryStorage{},
			&dataconnector.DataConnector{Name: "readonly", URL: "test://localhost:1234", ReadOnly: true},
		}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dataconnector.Add(tt.args.s, tt.args.m)
			assert.Equalf(t, tt.want, got, "Add() = %v, want %v", got, tt.want)
		})
	}
}
