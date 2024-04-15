package pull_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cgi-fr/lino/internal/infra/commonsql"
	infra "github.com/cgi-fr/lino/internal/infra/pull"
	"github.com/cgi-fr/lino/pkg/pull"

	"github.com/stretchr/testify/assert"
)

func TestCreateSelectPostgres(t *testing.T) {
	aTable := pull.Table{Name: "CUSTOMERS"}
	aFilter := pull.Filter{Limit: 5}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	defer db.Close()

	// Check SQL query is correctly created
	ds := infra.NewSQLDataSource("pg://server/name", "", nil, db, commonsql.PostgresDialect{})
	_, sql := ds.GetSelectSQLAndValues(aTable, aFilter)
	expectSQL := "SELECT * FROM \"CUSTOMERS\" WHERE  1=1  LIMIT 5"
	assert.Equal(t, expectSQL, sql)

	// Check SQL query can correctly excute in Postgres
	mock.ExpectQuery(sql).WillReturnRows()

	pgFactory := infra.NewPostgresDataSourceFactory()

	pgDS := pgFactory.New("pg://server/name", "")

	err = pgDS.(*infra.SQLDataSource).OpenWithDB(db)

	assert.Nil(t, err)

	_, err = pgDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}

func TestCreateSelectPostgresWithColumns(t *testing.T) {
	aTable := pull.Table{Name: "CUSTOMERS", Columns: []pull.Column{{Name: "Name"}, {Name: "Age"}}}
	aFilter := pull.Filter{Limit: 5}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	defer db.Close()

	// Check SQL query is correctly created
	ds := infra.NewSQLDataSource("pg://server/name", "", nil, db, commonsql.PostgresDialect{})
	_, sql := ds.GetSelectSQLAndValues(aTable, aFilter)
	expectSQL := "SELECT \"Name\", \"Age\" FROM \"CUSTOMERS\" WHERE  1=1  LIMIT 5"
	assert.Equal(t, expectSQL, sql)

	// Check SQL query can correctly excute in Postgres
	mock.ExpectQuery(sql).WillReturnRows()

	pgFactory := infra.NewPostgresDataSourceFactory()

	pgDS := pgFactory.New("pg://server/name", "")

	err = pgDS.(*infra.SQLDataSource).OpenWithDB(db)
	assert.Nil(t, err)

	_, err = pgDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}
