package model

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/api-x/internal/model/errors"
	"github.com/ryanreadbooks/whimer/misc/xstring"
)

type NoteId int64

// MarshalJSON implements the encoding json interface.
func (id NoteId) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return json.Marshal(nil)
	}

	result, err := infra.GetNoteIdObfuscate().Mix(int64(id))
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}


func (id *NoteId) fromBytes(data []byte) error {
	result, err := infra.GetNoteIdObfuscate().DeMix(xstring.FromBytes(data))
	if err != nil {
		return errors.ErrNoteNotFound
	}

	*id = NoteId(result)
	return nil
}

// for go-zero
func (id *NoteId) UnmarshalText(data []byte) error {
	return id.fromBytes(data)
}
