package load

// Mode to load rows
type Mode byte

const (
	// Truncate table before loading
	Truncate Mode = iota
	// Insert only new rows
	Insert
	// Upsert insert and update on conflict
	Upsert
	// Update only existing row
	Update
	end
)

// Modes
var modes = [...]string{
	"truncate",
	"insert",
	//"upsert",
	//"update",
}

// Modes list all modes string representation
func Modes() [2]string {
	return modes
}

// IsValidMode return true if value is a valide mode
func IsValidMode(value byte) bool {
	return value < byte(end)
}

// ParseMode return mode value of string representation of mode
func ParseMode(mode string) (Mode, *Error) {
	for i, m := range modes {
		if mode == m {
			return Mode(i), nil
		}
	}
	return end, &Error{mode + " is not a valide loading mode"}
}

// String representation
func (m Mode) String() string {
	for i, s := range modes {
		if Mode(i) == m {
			return s
		}
	}
	return "unknown"
}
