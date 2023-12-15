// Copyright (C) 2021 CGI France
//
// This file is part of LINO.
//
// LINO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LINO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LINO.  If not, see <http://www.gnu.org/licenses/>.

package analyse_test

import (
	"testing"

	"github.com/cgi-fr/lino/pkg/analyse"
	"github.com/cgi-fr/rimo/pkg/model"
	"github.com/stretchr/testify/assert"
)

type testDataSource struct{}

func (tds *testDataSource) Name() string {
	return "TestBase"
}

func (tds *testDataSource) ListTables() []string {
	return []string{"table1", "table2"}
}

func (tds *testDataSource) ListColumn(tableName string) []string {
	return []string{"col1", "col2"}
}

type testExtractor struct {
	values []interface{}
	index  int
}

func (tds *testExtractor) Open() error  { return nil }
func (tds *testExtractor) Close() error { return nil }

func (tds *testExtractor) ExtractValue() (bool, interface{}, error) {
	defer func() { tds.index++ }()

	if tds.index < len(tds.values) {
		return true, tds.values[tds.index], nil
	}

	return false, nil, nil
}

func (tds *testExtractor) New(tableName string, columnName string, limit uint) analyse.Extractor { //nolint:ireturn
	tds.index = 0
	return tds
}

type testWriter struct {
	result *model.Base
}

func (tw *testWriter) Write(report *model.Base) error {
	tw.result = report
	return nil
}

func TestAnalyseShouldNotReturnError(t *testing.T) {
	t.Parallel()

	dataSource := &testDataSource{}
	extractor := &testExtractor{values: []interface{}{nil, 1., 2., 3., 4., 5.}}
	writer := &testWriter{}
	driver := analyse.NewDriver(dataSource, extractor, writer, analyse.Config{Distinct: false})

	assert.NoError(t, driver.Analyse())

	assert.NotNil(t, writer.result)
	assert.Equal(t, "TestBase", writer.result.Name)
	assert.Equal(t, 2, len(writer.result.Tables))
	assert.Equal(t, uint(6), writer.result.Tables[0].Columns[0].MainMetric.Count)
	assert.Equal(t, uint(1), writer.result.Tables[0].Columns[0].MainMetric.Null)
	assert.Equal(t, uint(0), writer.result.Tables[0].Columns[0].MainMetric.Empty)
}
