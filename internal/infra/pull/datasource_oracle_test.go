package pull_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	infra "github.com/cgi-fr/lino/internal/infra/pull"
	"github.com/cgi-fr/lino/pkg/pull"

	"github.com/stretchr/testify/assert"
)

func TestCreateSelectOracle(t *testing.T) {
	aTable := pull.Table{Name: "CUSTOMERS"}
	aFilter := pull.Filter{Limit: 5}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT * FROM CUSTOMERS WHERE 1=1 AND rownum <= 5").WillReturnRows()

	pgFactory := infra.NewOracleDataSourceFactory()

	pgDS := pgFactory.New("pg://server/name", "")

	err = pgDS.(*infra.SQLDataSource).OpenWithDB(db)

	assert.Nil(t, err)

	_, err = pgDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}