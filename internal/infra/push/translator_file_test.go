package push_test

import (
	"strings"
	"testing"

	"github.com/cgi-fr/lino/internal/infra/push"
	driver "github.com/cgi-fr/lino/pkg/push"
	"github.com/stretchr/testify/assert"
)

type MemoryCache struct {
	reader *strings.Reader
}

func NewMemoryCache(content string) MemoryCache {
	return MemoryCache{reader: strings.NewReader(content)}
}

func (c MemoryCache) Close() error {
	return nil
}

func (c MemoryCache) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

func TestTranslator(t *testing.T) {
	translator := push.NewFileTranslator()

	cache := NewMemoryCache(`{"key":"key","value":"value"}`)

	error := translator.Load([]driver.Key{
		{TableName: "table1", ColumnName: "column1"},
		{TableName: "table1", ColumnName: "column2"},
		{TableName: "table2", ColumnName: "column1"},
		{TableName: "table2", ColumnName: "column2"},
	}, push.NewJSONRowIterator(cache))

	assert.Nil(t, error)
	assert.Equal(t, "key", translator.FindValue(driver.Key{TableName: "table1", ColumnName: "column1"}, "value"))
	assert.Equal(t, "key", translator.FindValue(driver.Key{TableName: "table1", ColumnName: "column2"}, "value"))
	assert.Equal(t, "key", translator.FindValue(driver.Key{TableName: "table2", ColumnName: "column1"}, "value"))
	assert.Equal(t, "key", translator.FindValue(driver.Key{TableName: "table2", ColumnName: "column2"}, "value"))
	assert.Equal(t, "value", translator.FindValue(driver.Key{TableName: "table3", ColumnName: "column1"}, "value"))
}

func TestPushWithNilValueDescriptor(t *testing.T) {
	oracleDialect := push.OracleDialect{}
	descriptor := push.ValueDescriptor{}
	cache := NewMemoryCache(`{"key":"key","value":"value"}`)
	iterator := push.NewJSONRowIterator(cache)
	for iterator.Next() {
		// row := iterator.Value()
		error := oracleDialect.ConvertValue("value", descriptor)
		assert.Nil(t, error)
	}
}
