package userchat

import (
	"encoding/json"
	"time"
)

func getAccurateTime() int64 {
	return time.Now().UnixNano()
}

func getNormalTime() int64 {
	return time.Now().Unix()
}

func makeBytes() []byte {
	return []byte{}
}

func makeJsonRawMessage() json.RawMessage {
	return json.RawMessage([]byte{})
}
