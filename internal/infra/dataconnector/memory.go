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
	"fmt"

	"github.com/cgi-fr/lino/pkg/dataconnector"
)

// NewMemoryStorage allocates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		repo: []dataconnector.DataConnector{},
	}
}

// MemoryStorage provides storage in memory
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
