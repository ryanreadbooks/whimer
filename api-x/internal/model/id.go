package model

import (
	"encoding/json"
	"strings"

	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/misc/xstring"
)

type NoteId uint64

// MarshalJSON implements the encoding json interface.
func (id NoteId) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return json.Marshal(nil)
	}

	result, err := infra.GetNoteIdObfuscate().MixU(uint64(id))
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// UnmarshalJSON implements the encoding json interface.
func (id *NoteId) UnmarshalJSON(data []byte) error {
	// convert null to 0
	if strings.TrimSpace(xstring.FromBytes(data)) == "null" {
		*id = 0
		return nil
	}

	// remove quotes
	if len(data) >= 2 {
		data = data[1 : len(data)-1]
	}

	result, err := infra.GetNoteIdObfuscate().DeMixU(xstring.FromBytes(data))
	if err != nil {
		return err
	}

	*id = NoteId(result)
	return nil
}

type SNoteId string

func (id SNoteId) Uint64() (uint64, error) {
	return infra.GetNoteIdObfuscate().DeMixU(string(id))
}
