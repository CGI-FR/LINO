package sequence

import (
	"strings"

	"github.com/rs/zerolog/log"
)

func Extract(e Updator, tables []Table, s Storage) *Error {
	seqs, err := e.Extract()
	if err != nil {
		return err
	}

	sequences := []Sequence{}

	for _, seq := range seqs {
		for _, tab := range tables {
			for _, key := range tab.Keys {
				if strings.Contains(seq, tab.Name) && strings.Contains(seq, key) {
					log.Debug().Str("table", tab.Name).Str("sequence", seq).Msg("Sequence - table match")

					sequences = append(sequences, Sequence{Name: seq, Table: tab.Name, Column: key})
				}
			}
		}
	}

	err = s.Store(sequences)
	if err != nil {
		return err
	}
	return nil
}

// Status retrun sequence with its current status
func Status(s Storage, u Updator) ([]Sequence, *Error) {
	sequences, err := s.List()
	if err != nil {
		return nil, err
	}

	result := []Sequence{}

	for _, seq := range sequences {
		seqUpdated, err := u.Status(seq)
		if err != nil {
			return nil, err
		}
		result = append(result, seqUpdated)
	}

	return result, nil
}

func Update(s Storage, u Updator) *Error {
	sequences, err := s.List()
	if err != nil {
		return err
	}

	err = u.Update(sequences)
	if err != nil {
		return err
	}

	return nil
}
