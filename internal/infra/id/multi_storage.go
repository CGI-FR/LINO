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

package id

import "makeit.imfr.cgi.com/lino/pkg/id"

// MultiStorage provides storage in multiple backend storages
type MultiStorage struct {
	storages []id.Storage
}

// NewMultiStorage create a new multi-storage
func NewMultiStorage(storages ...id.Storage) *MultiStorage {
	return &MultiStorage{
		storages: storages,
	}
}

// Store ingress descriptor in all the backends storages
func (s *MultiStorage) Store(adef id.IngressDescriptor) *id.Error {
	for _, s := range s.storages {
		err := s.Store(adef)
		if err != nil {
			return &id.Error{Description: err.Error()}
		}
	}
	return nil
}

func (s *MultiStorage) Read() (id.IngressDescriptor, *id.Error) {
	return s.storages[0].Read()
}
