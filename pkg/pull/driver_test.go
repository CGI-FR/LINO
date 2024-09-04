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

package pull_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/cgi-fr/jsonline/pkg/jsonline"
	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type Execution struct {
	Start  pull.Table
	Filter pull.Filter
	Result []string
}

type Test struct {
	DataSet    pull.DataSet
	Plan       pull.Plan
	Executions []Execution
}

func ToJSON(r pull.Row) string {
	row, _ := jsonline.NewTemplate().CreateRow(r)

	return row.String()
}

func LoadTest(filename string) (*Test, error) {
	yamlFile, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		return nil, fmt.Errorf(": %w", err)
	}

	// nolint: exhaustivestruct
	test := &Test{}

	err = yaml.Unmarshal(yamlFile, test)
	if err != nil {
		return nil, fmt.Errorf(": %w", err)
	}

	return test, nil
}

func RunTest(t *testing.T, test *Test) {
	t.Helper()
	// over.New(zerolog.New(os.Stderr))
	collector := pull.NewRowExporterCollector()

	puller := pull.NewPuller(test.Plan, pull.NewDataSourceInMemory(test.DataSet), collector, pull.NoTraceListener{})

	for _, execution := range test.Executions {
		collector.Reset()
		assert.NoError(t, puller.Pull(execution.Start, execution.Filter, nil, nil))
		assert.Len(t, collector.Result, len(execution.Result))

		for i := 0; i < len(execution.Result); i++ {
			assert.Equal(t, execution.Result[i], collector.Result[i].String())
		}
	}
}

func RunBench(b *testing.B, test *Test) {
	b.Helper()

	zerolog.SetGlobalLevel(zerolog.Disabled)

	collector := pull.NewRowExporterCollector()

	puller := pull.NewPuller(test.Plan, pull.NewDataSourceInMemory(test.DataSet), collector, pull.NoTraceListener{})

	for _, execution := range test.Executions {
		collector.Reset()
		assert.NoError(b, puller.Pull(execution.Start, execution.Filter, nil, nil))
		assert.Len(b, collector.Result, len(execution.Result))
	}
}

func LoadAndRunTest(t *testing.T, filename string) {
	t.Helper()

	test, err := LoadTest(filename)
	assert.NoError(t, err)

	RunTest(t, test)
}

func TestSimple(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "simple.yaml")
}

func TestSimpleSelect(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "simple_select.yaml")
}

func TestKeyNotSelected(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "key_not_selected.yaml")
}

func TestSimpleWithExport(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "simple_export.yaml")
}

func TestCycle(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "cycle.yaml")
}

func TestPull1(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "test1.yaml")
}

func TestPull2(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "test2.yaml")
}

func TestPull3(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "test3.yaml")
}

func TestPull4(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "test4.yaml")
}

func TestBug1(t *testing.T) {
	t.Parallel()

	LoadAndRunTest(t, "bug1.yaml")
}

func BenchmarkSimpleWithComponents(b *testing.B) {
	test, _ := LoadTest("simple.yaml")

	for n := 0; n < b.N; n++ {
		RunBench(b, test)
	}
}

func BenchmarkSimpleWithoutComponents(b *testing.B) {
	test, _ := LoadTest("simple.yaml")
	test.Plan.Components = map[pull.TableName]uint{}

	for n := 0; n < b.N; n++ {
		RunBench(b, test)
	}
}

func BenchmarkOverhead(b *testing.B) {
	test, _ := LoadTest("simple.yaml")
	test.Executions = []Execution{}

	for n := 0; n < b.N; n++ {
		RunBench(b, test)
	}
}

func BenchmarkOverheadWithoutComponents(b *testing.B) {
	test, _ := LoadTest("simple.yaml")
	test.Executions = []Execution{}
	test.Plan.Components = map[pull.TableName]uint{}

	for n := 0; n < b.N; n++ {
		RunBench(b, test)
	}
}
