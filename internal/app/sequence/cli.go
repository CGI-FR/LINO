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

package sequence

import (
	"fmt"
	"os"

	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/cgi-fr/lino/pkg/sequence"
	"github.com/cgi-fr/lino/pkg/table"
	"github.com/spf13/cobra"
)

var (
	dataconnectorStorage     dataconnector.Storage
	tableStorage             table.Storage
	sequenceStorage          sequence.Storage
	sequenceUpdatorFactories map[string]sequence.UpdatorFactory
)

// Inject dependencies
func Inject(dbas dataconnector.Storage, rs table.Storage, ss sequence.Storage, exmap map[string]sequence.UpdatorFactory) {
	dataconnectorStorage = dbas
	tableStorage = rs
	sequenceStorage = ss
	sequenceUpdatorFactories = exmap
}

// NewCommand implements the cli dataconnector command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sequence {extract|status|update} [arguments ...]",
		Short:   "Manage sequences",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s sequence extract mydatabase", fullName),
		Aliases: []string{"seq"},
	}
	cmd.AddCommand(newExtractCommand(fullName, err, out, in))
	cmd.AddCommand(newStatusCommand(fullName, err, out, in))
	cmd.AddCommand(newUpdateCommand(fullName, err, out, in))
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
