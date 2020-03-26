package pull

type table struct {
	name string
	pk   string
}

// NewTable initialize a new Table object
func NewTable(name string, pk string) Table {
	return table{name: name, pk: pk}
}

func (t table) Name() string       { return t.name }
func (t table) PrimaryKey() string { return t.pk }
func (t table) String() string     { return t.name }
