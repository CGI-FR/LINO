package id

import "makeit.imfr.cgi.com/lino/pkg/id"

// MultiStorage provides storage in multiple backend storages
type MultiStorage struct {
	storages []id.Storage
}

// NewMultiStorage create a new multi-storage
func NewMultiStorage(storages ...id.Storage) *MultiStorage {
	return &MultiStorage{
		storages: storages,
	}
}

// Store ingress descriptor in all the backends storages
func (s *MultiStorage) Store(adef id.IngressDescriptor) *id.Error {
	for _, s := range s.storages {
		err := s.Store(adef)
		if err != nil {
			return &id.Error{Description: err.Error()}
		}
	}
	return nil
}

func (s *MultiStorage) Read() (id.IngressDescriptor, *id.Error) {
	return s.storages[0].Read()
}
