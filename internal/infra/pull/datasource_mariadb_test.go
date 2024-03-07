package pull_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cgi-fr/lino/internal/infra/commonsql"
	infra "github.com/cgi-fr/lino/internal/infra/pull"
	"github.com/cgi-fr/lino/pkg/pull"

	"github.com/stretchr/testify/assert"
)

func TestCreateSelectMariadb(t *testing.T) {
	aTable := pull.Table{Name: "CUSTOMERS"}
	aFilter := pull.Filter{Limit: 5}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	defer db.Close()

	// Check SQL query is correctly created
	ds := infra.NewSQLDataSource("pg://server/name", "", nil, db, commonsql.MariadbDialect{})
	_, sql := ds.GetSelectSQLAndValues(aTable, aFilter)
	expectSQL := "SELECT * FROM `CUSTOMERS` WHERE  1=1   LIMIT 5"
	assert.Equal(t, expectSQL, sql)

	// Check SQL query can correctly excute in MariaDB
	mock.ExpectQuery(sql).WillReturnRows()

	pgFactory := infra.NewMariadbDataSourceFactory()

	pgDS := pgFactory.New("pg://server/name", "")

	err = pgDS.(*infra.SQLDataSource).OpenWithDB(db)

	assert.Nil(t, err)

	_, err = pgDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}

func TestCreateSelectMariadbWithColumns(t *testing.T) {
	columns := []pull.Column{
		{Name: "ID"},
		{Name: "Name"},
		{Name: "Age"},
	}
	aTable := pull.Table{Name: "CUSTOMERS", Columns: columns}
	aFilter := pull.Filter{Limit: 5}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	defer db.Close()

	// Check SQL query is correctly created
	ds := infra.NewSQLDataSource("pg://server/name", "", nil, db, commonsql.MariadbDialect{})
	_, sql := ds.GetSelectSQLAndValues(aTable, aFilter)
	expectSQL := "SELECT  `ID`,  `Name`,  `Age` FROM `CUSTOMERS` WHERE  1=1   LIMIT 5"
	assert.Equal(t, expectSQL, sql)

	// Check SQL query can correctly excute in MariaDB
	mock.ExpectQuery(sql).WillReturnRows()

	pgFactory := infra.NewMariadbDataSourceFactory()

	pgDS := pgFactory.New("pg://server/name", "")

	err = pgDS.(*infra.SQLDataSource).OpenWithDB(db)

	assert.Nil(t, err)

	_, err = pgDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}
