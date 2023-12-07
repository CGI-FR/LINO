package analyse

import "github.com/cgi-fr/lino/pkg/analyse"

func NewMapDataSource(name string, tables map[string][]string) analyse.DataSource {
	return &MapDataSource{name: name, tables: tables}
}

type MapDataSource struct {
	tables map[string][]string
	name   string
}

// ListColumn implements analyse.DataSource
func (ds *MapDataSource) ListColumn(tableName string) []string {
	return ds.tables[tableName]
}

// ListTables implements analyse.DataSource
func (ds *MapDataSource) ListTables() []string {
	result := []string{}
	for table := range ds.tables {
		result = append(result, table)
	}
	return result
}

// Name implements analyse.DataSource
func (ds *MapDataSource) Name() string {
	return ds.name
}
