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

	translator.Load([]driver.Key{
		{"table1", "column1"},
		{"table1", "column2"},
		{"table2", "column1"},
		{"table2", "column2"},
	}, push.NewJSONRowIterator(cache))

	assert.Equal(t, "key", translator.FindValue(driver.Key{"table1", "column1"}, "value"))
	assert.Equal(t, "key", translator.FindValue(driver.Key{"table1", "column2"}, "value"))
	assert.Equal(t, "key", translator.FindValue(driver.Key{"table2", "column1"}, "value"))
	assert.Equal(t, "key", translator.FindValue(driver.Key{"table2", "column2"}, "value"))
	assert.Equal(t, "value", translator.FindValue(driver.Key{"table3", "column1"}, "value"))
}
