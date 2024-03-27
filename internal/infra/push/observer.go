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

package push

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

type Observer struct {
	count int
	bar   *progressbar.ProgressBar
}

func NewObserver() *Observer {
	//nolint:gomnd
	pgb := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("Pushing ... "),
		progressbar.OptionSetItsString("entity"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowIts(),
		progressbar.OptionSpinnerType(11),
		progressbar.OptionThrottle(time.Millisecond*10),
		progressbar.OptionOnCompletion(func() { fmt.Fprintln(os.Stderr) }),
		// progressbar.OptionShowDescriptionAtLineEnd(),
	)

	return &Observer{
		count: 0,
		bar:   pgb,
	}
}

func (o *Observer) Pushed() {
	_ = o.bar.Add(1)

	o.count++

	o.bar.Describe(fmt.Sprintf("Pushed %d entities", o.count))
}

func (o *Observer) Close() {
	_ = o.bar.Close()
}
