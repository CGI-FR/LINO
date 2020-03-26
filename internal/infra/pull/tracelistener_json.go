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
