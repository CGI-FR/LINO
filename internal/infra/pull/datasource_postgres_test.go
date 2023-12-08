package pull_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

	mock.ExpectQuery("SELECT * FROM CUSTOMERS WHERE 1=1 LIMIT 5").WillReturnRows()

	pgFactory := infra.NewPostgresDataSourceFactory()

	pgDS := pgFactory.New("pg://server/name", "")

	err = pgDS.(*infra.SQLDataSource).OpenWithDB(db)

	assert.Nil(t, err)

	_, err = pgDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}

func TestCreateSelectPostgresWithColumns(t *testing.T) {
	aTable := pull.Table{Name: "CUSTOMERS", Columns: []pull.Column{{Name: "Name"}}}
	aFilter := pull.Filter{Limit: 5}

	// Créer une base de données simulée avec un mock
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	defer db.Close()

	// Définir l'expectation de la requête SELECT avec des colonnes spécifiques
	mock.ExpectQuery("SELECT Name FROM CUSTOMERS WHERE 1=1 LIMIT 5").WillReturnRows()

	// Créer une instance de la fabrique de source de données PostgreSQL
	pgFactory := infra.NewPostgresDataSourceFactory()

	// Créer une source de données PostgreSQL avec une URL factice
	pgDS := pgFactory.New("pg://server/name", "")

	// Ouvrir la source de données avec la base de données simulée
	err = pgDS.(*infra.SQLDataSource).OpenWithDB(db)
	assert.Nil(t, err)

	// Appeler RowReader avec la table et le filtre
	_, err = pgDS.RowReader(aTable, aFilter)
	assert.Nil(t, err)
}
