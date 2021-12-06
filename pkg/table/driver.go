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

package table

import "github.com/rs/zerolog/log"

// Extract table metadatas from a relational database.
func Extract(e Extractor, s Storage) *Error {
	tables, err := e.Extract()
	if err != nil {
		return err
	}
	err = s.Store(tables)
	if err != nil {
		return err
	}
	return nil
}

// Update sequence with the max of value + 1
func UpdateSequence(su Extractor, s Storage) *Error {
	tables, err := s.List()
	if err != nil {
		return err
	}

	for _, table := range tables {
		for _, seq := range table.Sequences {
			log.Debug().Str("sequence", seq.Name).Msg("update sequence")
			err := su.UpdateSequence(seq.Name, table.Name, seq.Key)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
