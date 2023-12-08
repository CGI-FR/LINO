package pull_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

	mock.ExpectQuery("SELECT  TOP 5 *  FROM CUSTOMERS  WHERE  1=1").WillReturnRows()

	msFactory := infra.NewSQLServerDataSourceFactory()

	msDS := msFactory.New("pg://server/name", "")

	err = msDS.(*infra.SQLDataSource).OpenWithDB(db)

	assert.Nil(t, err)

	_, err = msDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}
