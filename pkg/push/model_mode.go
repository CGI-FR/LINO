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

package push

// Mode to push rows
type Mode byte

const (
	// Truncate table before pushing
	Truncate Mode = iota
	// Insert only new rows
	Insert
	// Delete only existing row
	Delete
	// TODO Upsert insert and update on conflict
	// Update only existing row
	Update
	end
)

// Modes
var modes = [...]string{
	"truncate",
	"insert",
	"delete",
	// "upsert",
	"update",
}

// Modes list all modes string representation
func Modes() [4]string {
	return modes
}

// IsValidMode return true if value is a valide mode
func IsValidMode(value byte) bool {
	return value < byte(end)
}

// ParseMode return mode value of string representation of mode
func ParseMode(mode string) (Mode, *Error) {
	for i, m := range modes {
		if mode == m {
			return Mode(i), nil
		}
	}
	return end, &Error{mode + " is not a valide pushing mode"}
}

// String representation
func (m Mode) String() string {
	for i, s := range modes {
		if Mode(i) == m {
			return s
		}
	}
	return "unknown"
}
