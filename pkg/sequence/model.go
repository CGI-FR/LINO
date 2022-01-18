package sequence

type Sequence struct {
	Name   string
	Table  string
	Column string
	Value  int
}

type Error struct {
	Description string
}

type Table struct {
	Name string
	Keys []string
}
