package relation

// Extract relations from a relational database.
func Extract(e Extractor, s Storage) *Error {
	relations, err := e.Extract()
	if err != nil {
		return err
	}
	err = s.Store(relations)
	if err != nil {
		return err
	}
	return nil
}
