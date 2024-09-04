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

package pull

import (
	"fmt"

	"github.com/cgi-fr/lino/pkg/id"
	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/cgi-fr/lino/pkg/relation"
	"github.com/cgi-fr/lino/pkg/table"
	"github.com/rs/zerolog/log"
)

type builder struct {
	smap []id.Step
	rmap map[string]relation.Relation
	tmap map[string]table.Table

	exrmap map[string]pull.Relation
	extmap map[string]pull.Table
}

func newBuilder(plan id.PullerPlan, relations []relation.Relation, tables []table.Table) builder {
	smap := []id.Step{}
	for idx := uint(0); idx < plan.Len(); idx++ {
		smap = append(smap, plan.Step(idx))
	}
	rmap := map[string]relation.Relation{}
	for _, relation := range relations {
		rmap[relation.Name] = relation
	}
	tmap := map[string]table.Table{}
	for _, table := range tables {
		tmap[table.Name] = table
	}
	return builder{
		smap:   smap,
		rmap:   rmap,
		tmap:   tmap,
		exrmap: map[string]pull.Relation{},
		extmap: map[string]pull.Table{},
	}
}

func (b builder) plan() (pull.Plan, pull.Table, error) {
	log.Debug().Msg(fmt.Sprintf("there is %v step(s) to build", len(b.smap)))

	plan := pull.Plan{
		Relations:  []pull.Relation{},
		Components: map[pull.TableName]uint{},
	}

	for stepidx, step := range b.smap {
		log.Debug().Int("stepidx", stepidx+1).Stringer("step", step).Msg("building step")
		plan.Components[pull.TableName(step.Entry().Name())] = uint(stepidx)

		log.Trace().
			Int("stepidx", stepidx+1).
			Stringer("step", step).
			Msg(fmt.Sprintf("there is %v relation(s) to build for step %v", step.Relations().Len(), stepidx+1))

		if step.Following() != nil && step.Following().Name() != "" {
			result, err := b.buildRelation(step.Following())
			if err != nil {
				return pull.Plan{}, pull.Table{}, err
			}
			plan.Relations = append(plan.Relations, result...)
		}

		for relIdx := uint(0); relIdx < step.Relations().Len(); relIdx++ {
			result, err := b.buildRelation(step.Relations().Relation(relIdx))
			if err != nil {
				return pull.Plan{}, pull.Table{}, err
			}
			plan.Relations = append(plan.Relations, result...)
		}
	}

	startTable := b.getTable(b.smap[0].Entry().Name())

	log.Debug().Interface("startTable", startTable).Msg(fmt.Sprintf("finished building %v step(s) with success", len(b.smap)))
	return plan, startTable, nil
}

func (b builder) buildRelation(rel id.IngressRelation) ([]pull.Relation, error) {
	log.Trace().Stringer("rel", rel).Msg("building relation")

	result := []pull.Relation{}

	relyaml, ok := b.rmap[rel.Name()]
	if !ok {
		log.Error().Str("name", rel.Name()).Msg("missing relation in relations.yaml file")
		return nil, fmt.Errorf("missing relation in relations.yaml file : %v", rel.Name())
	}

	if rel.LookUpChild() {
		name := "many_" + rel.Name()
		exrel, ok := b.exrmap[name]
		if !ok {
			exrel = pull.Relation{
				Name:        pull.RelationName(rel.Name()),
				Cardinality: pull.Many,
				Local: pull.RelationTip{
					Table: b.getTable(rel.Parent().Name()),
					Keys:  relyaml.Parent.Keys,
				},
				Foreign: pull.RelationTip{
					Table: b.getTable(rel.Child().Name()),
					Keys:  relyaml.Child.Keys,
				},
				Where:  rel.WhereChild(),
				Select: rel.SelectChild(),
			}
		}
		b.exrmap[name] = exrel
		result = append(result, exrel)
	}

	if rel.LookUpParent() {
		name := "one_" + rel.Name()
		exrel, ok := b.exrmap[name]
		if !ok {
			exrel = pull.Relation{
				Name:        pull.RelationName(rel.Name()),
				Cardinality: pull.One,
				Foreign: pull.RelationTip{
					Table: b.getTable(rel.Parent().Name()),
					Keys:  relyaml.Parent.Keys,
				},
				Local: pull.RelationTip{
					Table: b.getTable(rel.Child().Name()),
					Keys:  relyaml.Child.Keys,
				},
				Where:  rel.WhereParent(),
				Select: rel.SelectParent(),
			}
		}
		b.exrmap[name] = exrel
		result = append(result, exrel)
	}

	return result, nil
}

func (b builder) getTable(name string) pull.Table {
	extable, ok := b.extmap[name]
	if !ok {
		table, ok := b.tmap[name]
		if !ok {
			log.Warn().Str("name", name).Msg("missing table in tables.yaml file")
			extable = pull.Table{
				Name:       pull.TableName(name),
				Keys:       []string{},
				Columns:    []pull.Column{},
				ExportMode: pull.ExportModeOnly,
			}
		} else {
			columns := []pull.Column{}
			for _, col := range table.Columns {
				columns = append(columns, pull.Column{
					Name:   col.Name,
					Export: col.Export,
				})
			}
			extable = pull.Table{
				Name:       pull.TableName(table.Name),
				Keys:       table.Keys,
				Columns:    columns,
				ExportMode: pull.ExportMode(table.ExportMode),
			}
		}
		b.extmap[name] = extable
	}
	return extable
}
