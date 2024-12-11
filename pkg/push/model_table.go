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

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cgi-fr/jsonline/pkg/jsonline"
	"github.com/rs/zerolog/log"
)

type table struct {
	name    string
	pk      []string
	columns ColumnList

	template  jsonline.Template
	filecache *FileCache
}

// NewTable initialize a new Table object
func NewTable(name string, pk []string, columns ColumnList) Table {
	return table{name: name, pk: pk, columns: columns, filecache: &FileCache{}}
}

func (t table) Name() string         { return t.name }
func (t table) PrimaryKey() []string { return t.pk }
func (t table) Columns() ColumnList  { return t.columns }
func (t table) String() string       { return t.name }

func (t table) GetColumn(name string) Column {
	if t.columns == nil {
		return nil
	}

	for idx := uint(0); idx < t.columns.Len(); idx++ {
		cur := t.columns.Column(idx)
		if name == cur.Name() {
			return cur
		}
	}

	return nil
}

type columnList struct {
	len   uint
	slice []Column
}

// NewColumnList initialize a new ColumnList object
func NewColumnList(columns []Column) ColumnList {
	return columnList{uint(len(columns)), columns}
}

func (l columnList) Len() uint              { return l.len }
func (l columnList) Column(idx uint) Column { return l.slice[idx] }
func (l columnList) String() string {
	switch l.len {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(l.slice[0])
	}
	sb := strings.Builder{}
	fmt.Fprintf(&sb, "%v", l.slice[0])
	for _, col := range l.slice[1:] {
		fmt.Fprintf(&sb, " -> %v", col)
	}
	return sb.String()
}

type column struct {
	name     string
	exp      string
	imp      string
	lgth     int64
	inbytes  bool
	truncate bool
}

// NewColumn initialize a new Column object
func NewColumn(name string, exp string, imp string, lgth int64, inbytes bool, truncate bool) Column {
	return column{name, exp, imp, lgth, inbytes, truncate}
}

func (c column) Name() string        { return c.name }
func (c column) Export() string      { return c.exp }
func (c column) Import() string      { return c.imp }
func (c column) Length() int64       { return c.lgth }
func (c column) LengthInBytes() bool { return c.inbytes }
func (c column) Truncate() bool      { return c.truncate }

type ImportedRow struct {
	jsonline.Row
}

func (t *table) applyFormats(formats map[string]string) {
	if t.columns == nil {
		return
	}

	if l := int(t.columns.Len()); l > 0 {
		columns := []Column{}
		for idx := 0; idx < l; idx++ {
			col := t.columns.Column(uint(idx))
			key := col.Name()

			if format, exist := formats[key]; exist {
				columns = append(columns, NewColumn(col.Name(), col.Export(), format, col.Length(), col.LengthInBytes(), col.Truncate()))
			} else {
				columns = append(columns, NewColumn(col.Name(), col.Export(), col.Import(), col.Length(), col.LengthInBytes(), col.Truncate()))
			}
		}
		t.columns = NewColumnList(columns)
	}
}

func (t *table) initTemplate() {
	t.template = jsonline.NewTemplate()

	if t.columns == nil {
		return
	}

	if l := int(t.columns.Len()); l > 0 {
		for idx := 0; idx < l; idx++ {
			col := t.columns.Column(uint(idx))
			key := col.Name()

			format, typ := parseFormatWithType(col.Import())

			if len(format) == 0 {
				log.Debug().Str("column", key).Msg("using value from export property for backward compatibility")
				format = col.Export() // backward compatibility with lino v2.1.0 and below
			}
			log.Debug().Str("column", key).Str("format", format).Str("typ", typ).Msg("parseFormatWithType")

			switch format {
			case "string", "file":
				t.template.WithMappedString(key, parseImportType(typ))
			case "numeric":
				t.template.WithMappedNumeric(key, parseImportType(typ))
			case "base64", "binary", "blob":
				t.template.WithMappedBinary(key, parseImportType(typ))
			case "datetime":
				t.template.WithMappedDateTime(key, parseImportType(typ))
			case "timestamp":
				t.template.WithMappedTimestamp(key, parseImportType(typ))
			case "no":
				t.template.WithHidden(key)
			default:
				if len(typ) > 0 {
					t.template.WithMappedAuto(key, parseImportType(typ))
				} else {
					log.Debug().Str("column", key).Str("typ", format).Msg("using value from import property as data type for backward compatibility")
					t.template.WithMappedAuto(key, parseImportType(format)) // backward compatibility with lino v2.1.0 and below
				}
			}
		}
	}
}

func (t table) Import(row map[string]interface{}) (ImportedRow, *Error) {
	if t.template == nil {
		t.initTemplate()
	}

	result := ImportedRow{t.template.CreateRowEmpty()}
	if err := result.Import(row); err != nil {
		return ImportedRow{}, &Error{Description: err.Error()}
	}

	if t.columns != nil {
		for idx := uint(0); idx < t.columns.Len(); idx++ {
			col := t.columns.Column(uint(idx))
			key := col.Name()

			format, _ := parseFormatWithType(col.Import())

			if value, exists := result.GetValue(key); format == "file" && exists && value.GetFormat() == jsonline.String && value.Raw() != nil {
				bytes, err := t.filecache.Load(result.GetString(key))
				if err != nil {
					return ImportedRow{}, &Error{Description: err.Error()}
				}
				result.SetValue(key, jsonline.NewValueAuto(bytes))
			} else if exists && col.Truncate() && col.Length() > 0 && value.GetFormat() == jsonline.String && result.GetOrNil(key) != nil {
				// autotruncate
				if col.LengthInBytes() {
					result.Set(key, truncateUTF8String(result.GetString(key), int(col.Length())))
				} else {
					result.Set(key, truncateRuneString(result.GetString(key), int(col.Length())))
				}
			}
		}
	}

	return result, nil
}

func parseImportType(exp string) jsonline.RawType {
	switch exp {
	case "int":
		return int(0)
	case "int64":
		return int64(0)
	case "int32":
		return int32(0)
	case "int16":
		return int16(0)
	case "int8":
		return int8(0)
	case "uint":
		return uint(0)
	case "uint64":
		return uint64(0)
	case "uint32":
		return uint32(0)
	case "uint16":
		return uint16(0)
	case "uint8":
		return uint8(0)
	case "float64":
		return float64(0)
	case "float32":
		return float32(0)
	case "bool":
		return false
	case "byte":
		return byte(0)
	case "rune":
		return rune(' ')
	case "string":
		return ""
	case "[]byte":
		return []byte{}
	case "time.Time":
		return time.Time{}
	case "json.Number":
		return json.Number("")
	default:
		return nil
	}
}

func parseFormatWithType(option string) (string, string) {
	parts := strings.Split(option, "(")
	if len(parts) != 2 {
		return option, ""
	}
	return parts[0], strings.Trim(parts[1], ")")
}

// truncateUTF8String truncate s to n bytes or less. If len(s) is more than n,
// truncate before the start of the first rune that doesn't fit. s should
// consist of valid utf-8.
func truncateUTF8String(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}

	return s[:n]
}

// truncateRuneString truncate s to n runes or less.
func truncateRuneString(s string, n int) string {
	if n <= 0 {
		return ""
	}

	if utf8.RuneCountInString(s) < n {
		return s
	}

	return string([]rune(s)[:n])
}
