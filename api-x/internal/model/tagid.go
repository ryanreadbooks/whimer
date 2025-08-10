package model

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/api-x/internal/model/errors"
	"github.com/ryanreadbooks/whimer/misc/xstring"
)

type TagId int64

func (id TagId) String() string {
	res, _ := infra.GetTagIdObfuscate().Mix(int64(id))
	return res
}

// MarshalJSON implements the encoding json interface.
func (id TagId) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return json.Marshal(nil)
	}

	result, err := infra.GetTagIdObfuscate().Mix(int64(id))
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (id *TagId) fromBytes(data []byte) error {
	result, err := infra.GetTagIdObfuscate().DeMix(xstring.FromBytes(data))
	if err != nil {
		return errors.ErrNoteNotFound
	}

	*id = TagId(result)
	return nil
}

// for go-zero
func (id *TagId) UnmarshalText(data []byte) error {
	return id.fromBytes(data)
}
