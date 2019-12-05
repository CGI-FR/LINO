package table

// Extract table metadatas from a relational database.
func Extract(e Extractor, s Storage) *Error {
	tables, err := e.Extract()
	if err != nil {
		return err
	}
	err = s.Store(tables)
	if err != nil {
		return err
	}
	return nil
}
