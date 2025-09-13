package id

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xstring"
)

type NoteId int64

var (
	noteIdObfuscate obfuscate.Obfuscate
)

func InitNoteIdObfuscate(opts ...obfuscate.Option) error {
	var err error
	noteIdObfuscate, err = obfuscate.NewConfuser(opts...)

	return err
}

func MustInitNoteIdObfuscate(opts ...obfuscate.Option) {
	if err := InitNoteIdObfuscate(opts...); err != nil {
		panic(err)
	}
}

func GetNoteIdObfuscate() obfuscate.Obfuscate {
	return noteIdObfuscate
}

func (n NoteId) String() string {
	result, _ := noteIdObfuscate.Mix(int64(n))
	return result
}

// MarshalJSON implements the encoding json interface.
func (id NoteId) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return json.Marshal("")
	}

	result, err := noteIdObfuscate.Mix(int64(id))
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (id *NoteId) fromBytes(data []byte) error {
	result, err := noteIdObfuscate.DeMix(xstring.FromBytes(data))
	if err != nil {
		return xerror.ErrArgs.Msg("note not found")
	}

	*id = NoteId(result)
	return nil
}

// for go-zero
func (id *NoteId) UnmarshalText(data []byte) error {
	return id.fromBytes(data)
}
