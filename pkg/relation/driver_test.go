package relation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"makeit.imfr.cgi.com/lino/pkg/relation"
)

// MemoryStorage provides storage of relations in memory
type MemoryStorage struct {
	repo []relation.Relation
}

// List relations stored in memory
func (s *MemoryStorage) List() ([]relation.Relation, *relation.Error) {
	return s.repo, nil
}

// Store relations in memory
func (s *MemoryStorage) Store(relations []relation.Relation) *relation.Error {
	s.repo = relations
	return nil
}

// MockStorage provides a mock for Storage interface
type MockStorage struct {
	fnList  func() ([]relation.Relation, *relation.Error)
	fnStore func(relations []relation.Relation) *relation.Error
}

// List relations stored
func (s *MockStorage) List() ([]relation.Relation, *relation.Error) {
	return s.fnList()
}

// Store relations
func (s *MockStorage) Store(relations []relation.Relation) *relation.Error {
	return s.fnStore(relations)
}

// MockExtractor mock Extractor interface
type MockExtractor struct {
	fn func() ([]relation.Relation, *relation.Error)
}

// Store a dataconnector in memory
func (e *MockExtractor) Extract() ([]relation.Relation, *relation.Error) {
	return e.fn()
}

func TestExtractEmptyDatabase(t *testing.T) {
	storage := &MemoryStorage{}
	Extractor := &MockExtractor{fn: func() ([]relation.Relation, *relation.Error) {
		return []relation.Relation{}, nil
	}}

	err := relation.Extract(Extractor, storage)

	assert.Nil(t, err, "An error occurred while using Add method")
	assert.Empty(t, storage.repo, "The relations storage should be empty")
}

func TestExtractNonEmptyDatabase(t *testing.T) {
	relation1 := relation.Relation{
		Name: "Relation1",
		Parent: relation.Table{
			Name: "Table1",
			Keys: []string{"Table1_key"},
		},
		Child: relation.Table{
			Name: "Table2",
			Keys: []string{"Table2_key"},
		},
	}
	storage := &MemoryStorage{}
	Extractor := &MockExtractor{fn: func() ([]relation.Relation, *relation.Error) {
		return []relation.Relation{relation1}, nil
	}}

	err := relation.Extract(Extractor, storage)

	assert.Nil(t, err, "An error occurred while using Add method")
	assert.Len(t, storage.repo, 1, "The relations storage should contains 1 relation")
	assert.ElementsMatch(t, storage.repo, []relation.Relation{relation1}, "Unexpected relations storage content")
}

func TestExtractorror(t *testing.T) {
	storage := &MemoryStorage{}
	Extractor := &MockExtractor{fn: func() ([]relation.Relation, *relation.Error) {
		return nil, &relation.Error{Description: "expected error"}
	}}

	err := relation.Extract(Extractor, storage)

	assert.NotNil(t, err, "An error should occur while using Extract method")
	assert.EqualError(t, err, "expected error")
}

func TestStoreError(t *testing.T) {
	relation1 := relation.Relation{
		Name: "Relation1",
		Parent: relation.Table{
			Name: "Table1",
			Keys: []string{"Table1_key"},
		},
		Child: relation.Table{
			Name: "Table2",
			Keys: []string{"Table2_key"},
		},
	}
	Extractor := &MockExtractor{fn: func() ([]relation.Relation, *relation.Error) {
		return []relation.Relation{relation1}, nil
	}}
	storage := &MockStorage{
		fnStore: func(relations []relation.Relation) *relation.Error {
			return &relation.Error{Description: "expected error"}
		},
	}

	err := relation.Extract(Extractor, storage)

	assert.NotNil(t, err, "An error should occur while using Extract method")
	assert.EqualError(t, err, "expected error")
}
