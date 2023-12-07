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

package analyse

import (
	"database/sql"
	"fmt"

	"github.com/cgi-fr/lino/pkg/analyse"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/xo/dburl"
)

// SQLExtractorFactory exposes methods to create new Postgres pullers.
type SQLExtractorFactory struct{}

// NewSQLExtractorFactory creates a new postgres datasource factory.
func NewSQLExtractorFactory() *SQLExtractorFactory {
	return &SQLExtractorFactory{}
}

// SQLDataSource to read in the analyse process.
type SQLDataSource struct {
	url    string
	schema string
	dbx    *sqlx.DB
	db     *sql.DB
}

// ExtractValues implements analyse.DataSource
func (ds *SQLDataSource) ExtractValues(tableName string, columnName string) ([]interface{}, error) {
	result := []interface{}{}

	log.Trace().Str("tablename", tableName).Str("columnname", columnName).Msg("extract values")

	err := ds.Open()
	if err != nil {
		log.Error().Err(err).Msg("Connection failed")
		return result, err
	}
	defer ds.db.Close()

	cursor, err := ds.db.Query(fmt.Sprintf("select %s from %s", columnName, tableName))
	if err != nil {
		log.Error().Err(err).Msg("SQL select failed")
		return result, err
	}
	for cursor.Next() {
		var value interface{}
		err = cursor.Scan(&value)
		if err != nil {
			log.Error().Err(err).Msg("SQL scan failed")
			return result, err
		}

		result = append(result, value)
	}
	return result, nil
}

func (e *SQLExtractorFactory) New(url string, schema string) analyse.Extractor {
	return &SQLDataSource{
		url:    url,
		schema: schema,
	}
}

// Open a connection to the SQL DB
func (ds *SQLDataSource) Open() error {
	db, err := dburl.Open(ds.url)
	if err != nil {
		return err
	}

	ds.db = db

	u, err := dburl.Parse(ds.url)
	if err != nil {
		return err
	}

	ds.dbx = sqlx.NewDb(db, u.UnaliasedDriver)

	err = ds.dbx.Ping()
	if err != nil {
		return err
	}

	return nil
}
