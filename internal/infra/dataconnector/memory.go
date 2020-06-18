package dataconnector

import (
	"fmt"

	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
)

// NewMemoryStorage allocates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		repo: []dataconnector.DataConnector{},
	}
}

// MemoryStorage provides storage in memory
type MemoryStorage struct {
	repo []dataconnector.DataConnector
}

// List all dataconnector stored in memory
func (s *MemoryStorage) List() ([]dataconnector.DataConnector, *dataconnector.Error) {
	return s.repo, nil
}

// Store a dataconnector in memory
func (s *MemoryStorage) Store(m *dataconnector.DataConnector) *dataconnector.Error {
	fmt.Println("Appending", m)
	s.repo = append(s.repo, *m)
	return nil
}
