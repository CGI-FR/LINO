package extract

import (
	"fmt"
	"strings"
)

type filter struct {
	limit  uint
	values Row
}

// NewFilter initialize a new Filter object
func NewFilter(limit uint, values Row) Filter {
	return filter{limit: limit, values: values}
}

func (f filter) Limit() uint { return f.limit }
func (f filter) Values() Row { return f.values }
func (f filter) String() string {
	builder := &strings.Builder{}
	cnt := len(f.Values())
	for key, value := range f.Values() {
		fmt.Fprintf(builder, "%v=%v", key, value)
		cnt--
		if cnt > 0 {
			fmt.Fprint(builder, " ")
		}
	}
	if len(f.Values()) == 0 && f.Limit() == 0 {
		fmt.Fprintf(builder, "true")
	}
	if f.Limit() > 0 {
		if len(f.Values()) > 0 {
			fmt.Fprint(builder, " ")
		}
		fmt.Fprintf(builder, "limit %v", f.Limit())
	}
	return builder.String()
}
