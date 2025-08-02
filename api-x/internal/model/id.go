package model

import (
	"encoding/json"
	"strings"

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

// UnmarshalJSON implements the encoding json interface.
func (id *NoteId) UnmarshalJSON(data []byte) error {
	// convert null to 0
	if strings.TrimSpace(xstring.FromBytes(data)) == "null" {
		*id = 0
		return nil
	}

	if len(data) > 2 {
		if data[0] == '"' && data[len(data)-1] == '"' {
			data = data[1 : len(data)-1]
		}
	}

	return id.fromBytes(data)
}

func (id *NoteId) UnmarshalText(data []byte) error {
	return id.fromBytes(data)
}
