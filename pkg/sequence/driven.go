package sequence

// UpdatorFactory exposes methods to create new Updators.
type UpdatorFactory interface {
	New(url string, schema string) Updator
}

// Updator allows to extract sequence from a relational database.
type Updator interface {
	Extract() ([]string, *Error)
	Status(sequence Sequence) (Sequence, *Error)
	Update([]Sequence) *Error
}

// Storage allows to store and retrieve sequences objects.
type Storage interface {
	List() ([]Sequence, *Error)
	Store(sequences []Sequence) *Error
}
