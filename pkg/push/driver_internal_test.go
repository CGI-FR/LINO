// NOTE: this test file introduces a unit test for combineErrors helper.
package push

import (
	"strings"
	"testing"
)

func TestCombineErrorsAggregatesMultipleErrors(t *testing.T) {
	e1 := &Error{Description: "first error"}
	e2 := &Error{Description: "second error"}

	combined := combineErrors(e1, e2, nil)
	if combined == nil {
		t.Fatal("expected combined error, got nil")
	}
	if !strings.Contains(combined.Description, "first error") || !strings.Contains(combined.Description, "second error") {
		t.Fatalf("combined Description did not contain both parts; got: %q", combined.Description)
	}
}

func TestCombineErrorsReturnsSingle(t *testing.T) {
	e1 := &Error{Description: "only one"}
	combined := combineErrors(nil, e1, nil)
	if combined == nil {
		t.Fatal("expected single error returned, got nil")
	}
	if combined != e1 {
		t.Fatalf("expected original *Error back, got different value: %v", combined)
	}
}
