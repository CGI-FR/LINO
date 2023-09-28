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
	"fmt"
	"testing"

	"github.com/cgi-fr/lino/pkg/analyse"
	"github.com/cgi-fr/rimo/pkg/model"
	"github.com/cgi-fr/rimo/pkg/rimo"
	"github.com/stretchr/testify/assert"
)

type rimoAnalyser struct {
	writer rimo.Writer
}

func (ra rimoAnalyser) Analyse(ds *analyse.ColumnIterator) error {
	return rimo.AnalyseBase(ds, ra)
}

func (ra rimoAnalyser) Export(base *model.Base) error {
	return ra.writer.Export(base)
}

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

func (tds *testDataSource) ExtractValues(columnName string) []interface{} {
	return []interface{}{1., 2., 3., 4., 5.}
}

type testWriter struct {
	result *model.Base
}

func (tw *testWriter) Export(report *model.Base) error {
	tw.result = report
	return nil
}

func TestAnalyseShouldNotReturnError(t *testing.T) {
	t.Parallel()
	dataSource := &testDataSource{}
	writer := &testWriter{}
	analyser := rimoAnalyser{writer}

	err := analyse.Do(dataSource, analyser)

	assert.Nil(t, err)
	assert.NotNil(t, writer.result)
	assert.Equal(t, "TestBase", writer.result.Name)
	assert.Equal(t, 2, len(writer.result.Tables))
	assert.Equal(t, 5, writer.result.Tables[0].Columns[0].MainMetric.Count)
}

func TestColumnIteratorNext(t *testing.T) {
	t.Parallel()

	dataSource := &testDataSource{}

	iterator := analyse.NewColumnIterator(dataSource)

	for table := 1; table < 3; table++ {
		for c := 1; c < 3; c++ {
			assert.True(t, iterator.Next())
			_, tableName, columnName, err := iterator.Value()
			assert.Nil(t, err)
			assert.Equal(t, fmt.Sprintf("table%d", table), tableName)
			assert.Equal(t, fmt.Sprintf("col%d", c), columnName)
		}
	}

	assert.False(t, iterator.Next())
}
