package relation

// ExtractorFactory exposes methods to create new extractors.
type ExtractorFactory interface {
	New(url string) Extractor
}

// Extractor allows to extract relations from a relational database.
type Extractor interface {
	Extract() ([]Relation, *Error)
}

// Storage allows to store and retrieve Relations objects.
type Storage interface {
	List() ([]Relation, *Error)
	Store(relations []Relation) *Error
}
