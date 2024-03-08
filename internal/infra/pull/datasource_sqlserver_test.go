package pull_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cgi-fr/lino/internal/infra/commonsql"
	infra "github.com/cgi-fr/lino/internal/infra/pull"
	"github.com/cgi-fr/lino/pkg/pull"

	"github.com/stretchr/testify/assert"
)

func TestCreateSelectSQLServer(t *testing.T) {
	aTable := pull.Table{Name: "CUSTOMERS"}
	aFilter := pull.Filter{Limit: 5}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	defer db.Close()

	// Check SQL query is correctly created
	ds := infra.NewSQLDataSource("pg://server/name", "", nil, db, commonsql.SQLServerDialect{})
	_, sql := ds.GetSelectSQLAndValues(aTable, aFilter)
	expectSQL := "SELECT TOP 5 * FROM CUSTOMERS WHERE  1=1 "
	assert.Equal(t, expectSQL, sql)

	// Check SQL query can correctly excute in SQL Server
	mock.ExpectQuery(sql).WillReturnRows()

	msFactory := infra.NewSQLServerDataSourceFactory()

	msDS := msFactory.New("pg://server/name", "")

	err = msDS.(*infra.SQLDataSource).OpenWithDB(db)

	assert.Nil(t, err)

	_, err = msDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}

func TestCreateSelectSQLServerWithColumns(t *testing.T) {
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
	ds := infra.NewSQLDataSource("pg://server/name", "", nil, db, commonsql.SQLServerDialect{})
	_, sql := ds.GetSelectSQLAndValues(aTable, aFilter)
	expectSQL := "SELECT TOP 5  ID, Name, Age FROM CUSTOMERS WHERE  1=1 "
	assert.Equal(t, expectSQL, sql)

	// Check SQL query can correctly excute in SQL Server
	mock.ExpectQuery(sql).WillReturnRows()

	msFactory := infra.NewSQLServerDataSourceFactory()

	msDS := msFactory.New("pg://server/name", "")

	err = msDS.(*infra.SQLDataSource).OpenWithDB(db)

	assert.Nil(t, err)

	_, err = msDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}
