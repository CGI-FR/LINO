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

func (ra rimoAnalyser) Analyse(ds analyse.DataSource) error {
	return rimo.AnalyseBase(ds, ra)
}

func (ra rimoAnalyser) Export(base *model.Base) error {
	return ra.writer.Export(base)
}

type testDataSource struct {
	collumnMax int
}

func (tds *testDataSource) BaseName() string {
	return "TestBase"
}

func (tds *testDataSource) Next() bool {
	tds.collumnMax--
	return tds.collumnMax > 0
}

func (tds *testDataSource) Value() ([]interface{}, string, string, error) {
	columnName := fmt.Sprintf("collumnName_%d", tds.collumnMax)
	return []interface{}{1., 2., 3., 4., 5.}, columnName, "tableName", nil
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
	dataSource := &testDataSource{collumnMax: 10}
	writer := &testWriter{}
	analyser := rimoAnalyser{writer}

	err := analyse.Do(dataSource, analyser)

	assert.Nil(t, err)
	assert.Equal(t, 0, dataSource.collumnMax)
	assert.NotNil(t, writer.result)
	assert.Equal(t, "TestBase", writer.result.Name)
	assert.Equal(t, 1, len(writer.result.Tables))
	assert.Equal(t, 5, writer.result.Tables[0].Columns[0].MainMetric.Count)
}
