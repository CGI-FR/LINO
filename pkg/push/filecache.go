// Copyright (C) 2024 CGI France
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

package push

import "os"

type FileCache struct {
	content map[string][]byte
}

func (fc *FileCache) Load(path string) ([]byte, error) {
	if bytes, ok := fc.content[path]; ok {
		return bytes, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, &Error{Description: err.Error()}
	}

	return bytes, nil
}
