package entity

import (
	whispervo "github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/vo"
)

type Chat struct {
	Id    string             `json:"id"`
	Name  string             `json:"name,omitempty"`
	Type  whispervo.ChatType `json:"type"`
	Ctime int64              `json:"ctime"`
}
