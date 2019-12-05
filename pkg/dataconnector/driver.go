package dataconnector

// Add an alias to the storage, if it does not exist
func Add(s Storage, m *DataConnector) *Error {
	exist, err := Get(s, m.Name)
	if err != nil {
		return err
	}
	if exist != nil {
		return nil
	}
	err = s.Store(m)
	if err != nil {
		return err
	}
	return nil
}

// Get an alias from the storage
func Get(s Storage, name string) (*DataConnector, *Error) {
	list, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, a := range list {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, nil
}

// List all stored aliases
func List(s Storage) ([]DataConnector, *Error) {
	aliases, err := s.List()
	if err != nil {
		return nil, err
	}
	if aliases == nil {
		aliases = []DataConnector{}
	}
	return aliases, err
}
