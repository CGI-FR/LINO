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

package dataconnector

import (
	"os"
	"strings"
	"testing"

	"github.com/cgi-fr/lino/pkg/dataconnector"
)

func TestYAMLStorageURLValueHolder(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Error(err)
		}
	})

	contents := `version: v1
dataconnectors:
  - name: literal
    url: postgresql://localhost/literal
  - name: value
    url:
      value: postgresql://localhost/value
  - name: environment
    url:
      valueFromEnv: DB_URL
`
	if err := os.WriteFile("dataconnector.yaml", []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	storage := NewYAMLStorage()
	connectors, connectorErr := storage.List()
	if connectorErr != nil {
		t.Fatal(connectorErr)
	}
	if len(connectors) != 3 {
		t.Fatalf("expected 3 connectors, got %d", len(connectors))
	}
	if connectors[0].URL.Value != "postgresql://localhost/literal" {
		t.Errorf("unexpected literal URL: %q", connectors[0].URL.Value)
	}
	if connectors[1].URL.Value != "postgresql://localhost/value" {
		t.Errorf("unexpected URL value: %q", connectors[1].URL.Value)
	}
	if connectors[2].URL.ValueFromEnv != "DB_URL" {
		t.Errorf("unexpected URL environment variable: %q", connectors[2].URL.ValueFromEnv)
	}

	connector := dataconnector.DataConnector{
		Name: "added",
		URL:  dataconnector.ValueHolder{ValueFromEnv: "ADDED_DB_URL"},
	}
	if connectorErr := storage.Store(&connector); connectorErr != nil {
		t.Fatal(connectorErr)
	}
	written, err := os.ReadFile("dataconnector.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "url: postgresql://localhost/literal") {
		t.Error("literal URL was not preserved as a YAML scalar")
	}
	if !strings.Contains(string(written), "valueFromEnv: ADDED_DB_URL") {
		t.Error("URL environment variable was not written")
	}
}
