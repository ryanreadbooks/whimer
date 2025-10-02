package v2

import (
	"encoding/json"

	msgpack "github.com/vmihailenco/msgpack/v5"
)

type Serializer interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
}

type JSONSerializer struct{}

func (j JSONSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (j JSONSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type MsgPackSerializer struct{}

func (j MsgPackSerializer) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (j MsgPackSerializer) Unmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}
