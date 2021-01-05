package localstorage

import (
	"strings"
)

// Errors raised by package.
var (
	ErrPermissionDenied  = constError("permission denied")
	ErrInvalidParameters = constError("invalid parameters")
	ErrMasterPassword    = constError("invalid master password")
	ErrInvalidStorage    = constError("invalid storage")
)

type constError string

func (err constError) Error() string {
	return string(err)
}
func (err constError) Is(target error) bool {
	ts := target.Error()
	es := string(err)
	return ts == es || strings.HasPrefix(ts, es+": ")
}
