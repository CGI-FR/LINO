package table

// ExtractorFactory exposes methods to create new extractors.
type ExtractorFactory interface {
	New(url string) Extractor
}

// Extractor allows to extract primary keys from a relational database.
type Extractor interface {
	Extract() ([]Table, *Error)
}

// Storage allows to store and retrieve Relations objects.
type Storage interface {
	List() ([]Table, *Error)
	Store(tables []Table) *Error
}
