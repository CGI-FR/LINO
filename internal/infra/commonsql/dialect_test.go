package commonsql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLServerDialect_Select(t *testing.T) {
	dialect := SQLServerDialect{}

	tableName := "MyTable"
	schemaName := "dbo"
	whereClause := "column1 = 1"
	distinct := true
	columns := []string{"column1", "column2"}
	expectedResult := "SELECT DISTINCT column1, column2 FROM dbo.MyTable WHERE column1 = 1"

	result := dialect.Select(tableName, schemaName, whereClause, distinct, columns...)

	assert.Equal(t, expectedResult, result)
}

func TestSQLServerDialect_SelectLimit(t *testing.T) {
	dialect := SQLServerDialect{}

	tableName := "MyTable"
	schemaName := "dbo"
	whereClause := "column1 = 1"
	distinct := false
	limit := uint(10)
	expectedResult := "SELECT TOP 10 * FROM dbo.MyTable WHERE column1 = 1"

	result := dialect.SelectLimit(tableName, schemaName, whereClause, distinct, limit)

	assert.Equal(t, expectedResult, result)
}

func TestSQLServerDialect_CreateSelect(t *testing.T) {
	dialect := SQLServerDialect{}

	sel := "SELECT"
	where := "WHERE column1 = 1"
	limit := "TOP 10"
	columns := "column1, column2"
	from := "FROM dbo.MyTable"
	expectedResult := "SELECT TOP 10 column1, column2 FROM dbo.MyTable WHERE column1 = 1"

	result := dialect.CreateSelect(sel, where, limit, columns, from)

	assert.Equal(t, expectedResult, result)
}

func TestPostgresDialect_Select(t *testing.T) {
	dialect := PostgresDialect{}

	tableName := "MyTable"
	schemaName := "dbo"
	whereClause := "column1 = 1"
	distinct := true
	columns := []string{"column1", "column2"}
	expectedResult := "SELECT DISTINCT  \"column1\", \"column2\" FROM \"dbo\".\"MyTable\" WHERE column1 = 1"

	result := dialect.Select(tableName, schemaName, whereClause, distinct, columns...)

	assert.Equal(t, expectedResult, result)
}

func TestPostgresDialect_SelectLimit(t *testing.T) {
	dialect := PostgresDialect{}

	tableName := "MyTable"
	schemaName := "dbo"
	whereClause := "column1 = 1"
	distinct := false
	limit := uint(10)
	expectedResult := "SELECT  \"column1\", \"column2\" FROM \"dbo\".\"MyTable\" WHERE column1 = 1 LIMIT 10"
	columns := []string{"column1", "column2"}

	result := dialect.SelectLimit(tableName, schemaName, whereClause, distinct, limit, columns...)

	assert.Equal(t, expectedResult, result)
}
