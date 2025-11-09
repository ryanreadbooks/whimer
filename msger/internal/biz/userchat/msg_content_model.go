package userchat

import (
	"encoding/json"
	"fmt"

	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

var (
	ErrUnsupportedMsgContent = fmt.Errorf("unsupported msg content")
)

type MsgContent interface {
	Bytes() ([]byte, error)
	MsgType() model.MsgType
	Preview() string
	Parse([]byte) (MsgContent, error)
}

// 解析出msgcontent
func ParseMsgContent(b []byte) (MsgContent, model.MsgType, error) {
	for t, candidate := range msgContentPool {
		ret, err := candidate.Parse(b)
		if err == nil {
			return ret, t, nil
		}
	}

	return nil, model.MsgTypeUnknown, ErrUnsupportedMsgContent
}

var (
	_ MsgContent = &MsgContentText{}
	_ MsgContent = &MsgContentImage{}

	msgContentPool = map[model.MsgType]MsgContent{
		model.MsgText:  &MsgContentText{},
		model.MsgImage: &MsgContentImage{},
	}
)

// 纯文本
type MsgContentText struct {
	Text string `json:"t"`
}

func (c *MsgContentText) Bytes() ([]byte, error) {
	return json.Marshal(c)
}

func (c *MsgContentText) MsgType() model.MsgType {
	return model.MsgText
}

func (c *MsgContentText) Preview() string {
	return c.Text
}

func (c *MsgContentText) Parse(b []byte) (MsgContent, error) {
	var cc MsgContentText
	err := json.Unmarshal(b, &cc)
	if err != nil {
		return nil, err
	}

	return &cc, nil
}

// 纯图片
type MsgContentImage struct {
	Key    string `json:"k"`
	Format string `json:"f"`
	Width  uint32 `json:"w"`
	Height uint32 `json:"h"`
}

func (c *MsgContentImage) Bytes() ([]byte, error) {
	return json.Marshal(c)
}

func (c *MsgContentImage) MsgType() model.MsgType {
	return model.MsgImage
}

func (c *MsgContentImage) Preview() string {
	return "[图片]"
}

func (c *MsgContentImage) Parse(b []byte) (MsgContent, error) {
	var cc MsgContentImage
	err := json.Unmarshal(b, &cc)
	if err != nil {
		return nil, err
	}

	return &cc, nil
}
