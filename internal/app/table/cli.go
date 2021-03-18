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

package table

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
	"makeit.imfr.cgi.com/lino/pkg/table"
)

var dataconnectorStorage dataconnector.Storage
var tableStorage table.Storage
var tableExtractorFactories map[string]table.ExtractorFactory

// Inject dependencies
func Inject(dbas dataconnector.Storage, rs table.Storage, exmap map[string]table.ExtractorFactory) {
	dataconnectorStorage = dbas
	tableStorage = rs
	tableExtractorFactories = exmap
}

// NewCommand implements the cli dataconnector command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "table {extract} [arguments ...]",
		Short:   "Manage tables",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s table extract mydatabase", fullName),
		Aliases: []string{"tab"},
	}
	cmd.AddCommand(newExtractCommand(fullName, err, out, in))
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
