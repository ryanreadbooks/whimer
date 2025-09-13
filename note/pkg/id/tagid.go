package id

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xstring"
)

type TagId int64

var (
	tagIdObfuscate obfuscate.Obfuscate
)

func InitTagIdObfuscate(opts ...obfuscate.Option) error {
	var err error
	tagIdObfuscate, err = obfuscate.NewConfuser(opts...)

	return err
}

func MustTagIdObfuscate(opts ...obfuscate.Option) {
	if err := InitNoteIdObfuscate(opts...); err != nil {
		panic(err)
	}
}

func GetTagIdObfuscate() obfuscate.Obfuscate {
	return tagIdObfuscate
}

func (id TagId) String() string {
	res, _ := tagIdObfuscate.Mix(int64(id))
	return res
}

// MarshalJSON implements the encoding json interface.
func (id TagId) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return json.Marshal("")
	}

	result, err := tagIdObfuscate.Mix(int64(id))
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (id *TagId) fromBytes(data []byte) error {
	result, err := tagIdObfuscate.DeMix(xstring.FromBytes(data))
	if err != nil {
		return xerror.ErrArgs.Msg("tag not found")
	}

	*id = TagId(result)
	return nil
}

// for go-zero
func (id *TagId) UnmarshalText(data []byte) error {
	return id.fromBytes(data)
}
