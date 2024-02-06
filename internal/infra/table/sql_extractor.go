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
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/cgi-fr/lino/pkg/table"
	"github.com/xo/dburl"
)

// SQLExtractor provides table extraction logic from SQL database.
type SQLExtractor struct {
	url     string
	schema  string
	dialect Dialect
}

type Dialect interface {
	SQL(schema string) string
}

// NewSQLExtractor creates a new SQL extractor.
func NewSQLExtractor(url string, schema string, dialect Dialect) *SQLExtractor {
	return &SQLExtractor{
		url:     url,
		schema:  schema,
		dialect: dialect,
	}
}

// Extract tables from the database.
func (e *SQLExtractor) Extract() ([]table.Table, *table.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	SQL := e.dialect.SQL(e.schema)

	rows, err := db.Query(SQL)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	tables := []table.Table{}

	var (
		tableSchema string
		tableName   string
		keyColumns  string
	)

	for rows.Next() {
		err := rows.Scan(&tableSchema, &tableName, &keyColumns)
		if err != nil {
			return nil, &table.Error{Description: err.Error()}
		}
		ColumnInfo(db, tableName)
		table := table.Table{
			Name: tableName,
			Keys: strings.Split(keyColumns, ","),
		}
		tables = append(tables, table)
	}
	err = rows.Err()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	return tables, nil
}

func (e *SQLExtractor) Count(tableName string) (int, *table.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}

	SQL := `SELECT COUNT(*) FROM ` + tableName

	rows, err := db.Query(SQL)
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}

	var count int

	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, &table.Error{Description: err.Error()}
		}
	}

	err = rows.Err()
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}

	return count, nil
}

func ColumnInfo(db *sql.DB, tableName string) {
	// Exécution de la requête pour obtenir les informations sur les colonnes
	rows, err := db.Query("SELECT * FROM " + tableName + " LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Récupération des informations sur les colonnes
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		rows.Close()
		log.Fatal(err)
	}

	// Parcours des informations sur les colonnes
	for _, ct := range columnTypes {
		// Récupération du nom de la colonne
		columnName := ct.Name()

		// Récupération du type de données de la colonne
		dataType := ct.DatabaseTypeName()

		// Récupération de la longueur ou de la taille de la colonne (si applicable)
		columnLength, _ := ct.Length()
		columnSize, _ := ct.Length()

		// Affichage des informations sur la colonne
		fmt.Printf("Column Name: %s, Data Type: %s", columnName, dataType)
		if columnLength > 0 {
			fmt.Printf(", Length: %d", columnLength)
		} else if columnSize > 0 {
			fmt.Printf(", Size: %d", columnSize)
		}
		fmt.Println()
	}
}
