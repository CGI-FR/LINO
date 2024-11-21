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

	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/analyse"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/xo/dburl"
)

// SQLExtractorFactory exposes methods to create new Postgres pullers.
type SQLExtractorFactory struct {
	dialect commonsql.Dialect
}

// NewPostgresExtractorFactory creates a new postgres datasource factory.
func NewPostgresExtractorFactory() SQLExtractorFactory {
	return SQLExtractorFactory{
		dialect: commonsql.PostgresDialect{},
	}
}

// NewOracleExtractorFactory creates a new postgres datasource factory.
func NewOracleExtractorFactory() SQLExtractorFactory {
	return SQLExtractorFactory{
		dialect: commonsql.OracleDialect{},
	}
}

// NewMariaDBExtractorFactory creates a new postgres datasource factory.
func NewMariaDBExtractorFactory() SQLExtractorFactory {
	return SQLExtractorFactory{
		dialect: commonsql.MariadbDialect{},
	}
}

// NewDB2ExtractorFactory creates a new postgres datasource factory.
func NewDB2ExtractorFactory() SQLExtractorFactory {
	return SQLExtractorFactory{
		dialect: commonsql.Db2Dialect{},
	}
}

// NewSQLServerExtractorFactory creates a new postgres datasource factory.
func NewSQLServerExtractorFactory() SQLExtractorFactory {
	return SQLExtractorFactory{
		dialect: commonsql.SQLServerDialect{},
	}
}

func (e SQLExtractorFactory) New(url string, schema string) analyse.ExtractorFactory {
	return &SQLExtractor{
		url:     url,
		schema:  schema,
		dialect: e.dialect,
	}
}

type SQLExtractor struct {
	url     string
	schema  string
	dialect commonsql.Dialect
}

func (s SQLExtractor) New(tableName string, columnName string, limit uint, where string) analyse.Extractor { //nolint:ireturn
	return &SQLDataSource{
		url:     s.url,
		schema:  s.schema,
		table:   tableName,
		column:  columnName,
		limit:   limit,
		where:   where,
		dialect: s.dialect,
		dbx:     nil,
		db:      nil,
		cursor:  nil,
	}
}

// SQLDataSource to read in the analyse process.
type SQLDataSource struct {
	url     string
	schema  string
	table   string
	column  string
	limit   uint
	where   string
	dialect commonsql.Dialect
	dbx     *sqlx.DB
	db      *sql.DB
	cursor  *sql.Rows
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

	// Get WHERE Clause query
	sqlWhere, values := commonsql.GetWhereSQLAndValues(map[string]any{}, ds.where, ds.dialect)

	sql := ds.dialect.Select(ds.table, ds.schema, sqlWhere, false, commonsql.ColumnExportDefinition{Name: ds.column})

	// If log level is more than debug level, this function will log all SQL Query
	commonsql.LogSQLQuery(sql, values, ds.dialect)

	ds.cursor, err = ds.db.Query(sql)
	if err != nil {
		log.Error().Err(err).Msg("SQL select failed")
		return err
	}

	return nil
}

// Close a connection to the SQL DB
func (ds *SQLDataSource) Close() error {
	return ds.db.Close()
}

// ExtractValues implements analyse.DataSource
func (ds *SQLDataSource) ExtractValue() (bool, interface{}, error) {
	if ds.cursor.Next() {
		var value interface{}
		if err := ds.cursor.Scan(&value); err != nil {
			log.Error().Err(err).Msg("SQL scan failed")
			return false, nil, err
		}

		log.Trace().
			Str("tablename", ds.table).
			Str("columnname", ds.column).
			Interface("value", value).Msg("extract value")

		return true, value, nil
	}

	return false, nil, nil
}
