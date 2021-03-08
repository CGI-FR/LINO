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
	"encoding/json"
	"fmt"
	"os"
	"time"

	"makeit.imfr.cgi.com/lino/pkg/pull"
)

// JSONTraceListener JSON event tracer
type JSONTraceListener struct {
	last time.Time
	file *os.File
}

// NewJSONTraceListener create JsonTraceListner
func NewJSONTraceListener(file *os.File) JSONTraceListener {
	return JSONTraceListener{last: time.Now(), file: file}
}

// Event is catch by json tracer
type Event struct {
	Duration int64  `json:"duration"`
	Index    uint   `json:"index"`
	Entry    string `json:"entry"`
	Follow   string `json:"follow"`
	Filter   string `json:"filter"`
}

// TraceStep catch Step event.
func (t JSONTraceListener) TraceStep(s pull.Step, filter pull.Filter) pull.TraceListener {
	now := time.Now()
	event := Event{
		Duration: now.Sub(t.last).Milliseconds(),
		Index:    s.Index(),
		Entry:    s.Entry().Name(),
		Follow: func() string {
			if s.Follow() != nil {
				return s.Follow().Name()
			}
			return ""
		}(),
		Filter: fmt.Sprintf("%v", filter.Values()),
	}
	t.last = now
	jsonString, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(t.file, string(jsonString))
	return t
}
