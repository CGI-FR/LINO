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

import (
	"fmt"
	"sort"
)

// Extract table metadatas from a relational database.
func Extract(e Extractor, s Storage, onlyTables bool) *Error {
	tables, err := e.Extract()
	if err != nil {
		return err
	}

	if !onlyTables {
		SortKeysByColumnOrder(tables)
	}

	err = s.Store(tables)
	if err != nil {
		return err
	}
	return nil
}

// Count ligne in table `tableName`
func Count(s Storage, e Extractor) ([]TableCount, *Error) {
	tables, err := s.List()
	if err != nil {
		return nil, err
	}

	result := []TableCount{}
	for _, table := range tables {
		count, err := e.Count(table.Name)
		if err != nil {
			return nil, err
		}

		result = append(result, TableCount{table, count})
	}
	return result, nil
}

// AddOrUpdateColumn will update table definitions with given export and import types, it will add the column if necessary
func AddOrUpdateColumn(s Storage, tableName, columnName, exportType, importType string) (int, *Error) {
	tables, err := s.List()
	if err != nil {
		return 0, err
	}

	count := 0

	updatedTables := []Table{}
	for _, table := range tables {
		if table.Name == tableName {
			updatedTables = append(updatedTables, addOrUpdateColumn(table, columnName, exportType, importType))
			count++
		} else {
			updatedTables = append(updatedTables, table)
		}
	}

	if count == 0 {
		return 0, &Error{Description: fmt.Sprintf("there is no table named %v", tableName)}
	}

	if err := s.Store(updatedTables); err != nil {
		return count, err
	}

	return count, nil
}

func addOrUpdateColumn(table Table, columnName, exportType, importType string) Table {
	count := 0

	updatedColumns := []Column{}

	for _, column := range table.Columns {
		if column.Name == columnName {
			exportUpdate := column.Export
			if exportType != "" {
				exportUpdate = exportType
			}
			importUpdate := column.Import
			if importType != "" {
				importUpdate = importType
			}
			updatedColumns = append(updatedColumns, Column{
				Name:   columnName,
				Export: exportUpdate,
				Import: importUpdate,
			})
			count++
		} else {
			updatedColumns = append(updatedColumns, column)
		}
	}

	if count == 0 {
		updatedColumns = append(updatedColumns, Column{
			Name:   columnName,
			Export: exportType,
			Import: importType,
		})
	}

	table.Columns = updatedColumns

	return table
}

// RemoveColumn will update table definitions removing specified column
func RemoveColumn(s Storage, tableName, columnName string) (int, *Error) {
	tables, err := s.List()
	if err != nil {
		return 0, err
	}

	count := 0

	updatedTables := []Table{}
	for _, table := range tables {
		if table.Name == tableName {
			updatedTables = append(updatedTables, removeColumn(table, columnName))
			count++
		} else {
			updatedTables = append(updatedTables, table)
		}
	}

	if count == 0 {
		return 0, &Error{Description: fmt.Sprintf("there is no table named %v", tableName)}
	}

	if err := s.Store(updatedTables); err != nil {
		return count, err
	}

	return count, nil
}

func removeColumn(table Table, columnName string) Table {
	count := 0

	updatedColumns := []Column{}

	for _, column := range table.Columns {
		if column.Name == columnName {
			count++
		} else {
			updatedColumns = append(updatedColumns, column)
		}
	}

	table.Columns = updatedColumns

	return table
}

func SortKeysByColumnOrder(tables []Table) []Table {
	for _, table := range tables {
		// No need to sort table which only contain one or zero key
		if len(table.Keys) > 1 {
			// Create a map to match column names to their index in the list
			columnIndex := make(map[string]int)
			for i, column := range table.Columns {
				columnIndex[column.Name] = i
			}

			// Sort keys based on the order of column names
			sort.Slice(table.Keys, func(i, j int) bool {
				return columnIndex[table.Keys[i]] < columnIndex[table.Keys[j]]
			})
		}
	}

	return tables
}
