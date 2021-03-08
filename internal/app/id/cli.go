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

package id

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"makeit.imfr.cgi.com/lino/pkg/id"
	"makeit.imfr.cgi.com/lino/pkg/relation"
)

var idStorage id.Storage
var relStorage relation.Storage
var idExporter id.Exporter
var idJSONExporter id.Storage

// SetLogger if needed, default no logger
func SetLogger(l id.Logger) {
	id.SetLogger(l)
}

// Inject dependencies
func Inject(ids id.Storage, rels relation.Storage, ex id.Exporter, jSONEx id.Storage) {
	idStorage = ids
	relStorage = rels
	idExporter = ex
	idJSONExporter = jSONEx
}

// NewCommand implements the cli id command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "id {create,display-plan,show-graph,export,set-start-table,set-child-lookup,set-parent-lookup} [arguments ...]",
		Short:   "Manage ingress descriptor",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id create mydatabase public.customer", fullName),
	}
	cmd.AddCommand(newCreateCommand(fullName, err, out, in))
	cmd.AddCommand(newDisplayPlanCommand(fullName, err, out, in))
	cmd.AddCommand(newShowGraphCommand(fullName, err, out, in))
	cmd.AddCommand(newExportCommand(fullName, err, out, in))
	cmd.AddCommand(newSetStartTableCommand(fullName, err, out, in))
	cmd.AddCommand(newSetChildLookupCommand(fullName, err, out, in))
	cmd.AddCommand(newSetParentLookupCommand(fullName, err, out, in))
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
