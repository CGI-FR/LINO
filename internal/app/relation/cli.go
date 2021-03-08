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

package relation

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
	"makeit.imfr.cgi.com/lino/pkg/relation"
)

var dataconnectorStorage dataconnector.Storage
var relationStorage relation.Storage
var relationExtractorFactories map[string]relation.ExtractorFactory

// Inject dependencies
func Inject(dbas dataconnector.Storage, rs relation.Storage, exmap map[string]relation.ExtractorFactory) {
	dataconnectorStorage = dbas
	relationStorage = rs
	relationExtractorFactories = exmap
}

// NewCommand implements the cli dataconnector command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "relation {extract} [arguments ...]",
		Short:   "Manage relations",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s relation extract mydatabase", fullName),
		Aliases: []string{"rel"},
	}
	cmd.AddCommand(newExtractCommand(fullName, err, out, in))
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
