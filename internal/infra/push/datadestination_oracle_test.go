package push

import (
	"strings"
	"testing"

	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/push"
	_ "github.com/sijms/go-ora/v2"
	"github.com/stretchr/testify/assert"
)

func QuotedOracle(d SQLDialect, name string) string {
	return d.Quote(name)
}

func TestAppendColumnToSQLWithPreserveNothing(t *testing.T) {
	column := ValueDescriptor{
		name: "column",
		column: push.NewColumn(
			"column",
			"",
			"",
			0,
			false,
			false,

			push.PreserveNothing,
		),
	}

	sql := &strings.Builder{}
	d := OracleDialect{innerDialect: commonsql.OracleDialect{}}
	index := 0
	err := appendColumnToSQL(column, sql, d, index)
	if err != nil { // should not return an error
		t.Errorf("Expected no error, got %v", err)
	}
	expectedSQL := QuotedOracle(d, "column") + "=:v1"
	assert.Equal(t, expectedSQL, sql.String())
}

func TestAppendColumnToSQLWithPreserveBlank(t *testing.T) {
	column := ValueDescriptor{
		name: "column",
		column: push.NewColumn(
			"column",
			"",
			"",
			0,
			false,
			false,

			push.PreserveBlank,
		),
	}

	sql := &strings.Builder{}
	d := OracleDialect{innerDialect: commonsql.OracleDialect{}}
	index := 0
	err := appendColumnToSQL(column, sql, d, index)
	if err != nil { // should not return an error
		t.Errorf("Expected no error, got %v", err)
	}
	q := QuotedOracle(d, "column")
	expectedSQL := q +
		" = CASE WHEN " + q +
		" IS NULL THEN " + q +
		" WHEN TRIM(" + q +
		") IS NULL THEN " + q +
		" ELSE :v1 END"

	assert.Equal(t, expectedSQL, sql.String())
}

func TestAppendColumnToSQLWithPreserveEmpty(t *testing.T) {
	column := ValueDescriptor{
		name: "column",
		column: push.NewColumn(
			"column",
			"",
			"",
			0,
			false,
			false,

			push.PreserveEmpty,
		),
	}

	sql := &strings.Builder{}
	d := OracleDialect{innerDialect: commonsql.OracleDialect{}}
	index := 0
	err := appendColumnToSQL(column, sql, d, index)
	assert.NotNil(t, err)
}

func TestAppendColumnToSQLWithPreserveNull(t *testing.T) {
	column := ValueDescriptor{
		name: "column",
		column: push.NewColumn(
			"column",
			"",
			"",
			0,
			false,
			false,

			push.PreserveNull,
		),
	}

	sql := &strings.Builder{}
	d := OracleDialect{innerDialect: commonsql.OracleDialect{}}
	index := 0
	err := appendColumnToSQL(column, sql, d, index)
	if err != nil { // should not return an error
		t.Errorf("Expected no error, got %v", err)
	}
	q := QuotedOracle(d, "column")
	expectedSQL := q +
		" = CASE WHEN " + q +
		" IS NOT NULL THEN :v1 ELSE " + q +
		" END"
	assert.Equal(t, expectedSQL, sql.String())
}
