package pull_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	infra "github.com/cgi-fr/lino/internal/infra/pull"
	"github.com/cgi-fr/lino/pkg/pull"

	"github.com/stretchr/testify/assert"
)

func TestCreateSelectDb2(t *testing.T) {
	aTable := pull.Table{Name: "CUSTOMERS"}
	aFilter := pull.Filter{Limit: 5}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT  *  FROM CUSTOMERS  WHERE  1=1   FETCH FIRST 5 ROWS ONLY").WillReturnRows()

	db2Factory := infra.NewDb2DataSourceFactory()

	pgDS := db2Factory.New("pg://server/name", "")

	err = pgDS.(*infra.SQLDataSource).OpenWithDB(db)

	assert.Nil(t, err)

	_, err = pgDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}
