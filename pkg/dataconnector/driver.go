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

var logger Logger = Nologger{}

// Add an alias to the storage, if it does not exist
func Add(s Storage, m *DataConnector) *Error {
	exist, err := Get(s, m.Name)
	if err != nil {
		logger.Error(err.Description)
		return err
	}

	if exist != nil {
		return nil
	}
	err = s.Store(m)
	if err != nil {
		logger.Error(err.Description)
		return err
	}
	return nil
}

// Get an alias from the storage
func Get(s Storage, name string) (*DataConnector, *Error) {
	list, err := s.List()
	if err != nil {
		logger.Error(err.Description)
		return nil, err
	}
	for _, a := range list {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, nil
}

// List all stored aliases
func List(s Storage) ([]DataConnector, *Error) {
	aliases, err := s.List()
	if err != nil {
		logger.Error(err.Description)
		return nil, err
	}
	if aliases == nil {
		aliases = []DataConnector{}
	}
	return aliases, err
}

// SetLogger if needed, default no logger
func SetLogger(l Logger) {
	logger = l
}
