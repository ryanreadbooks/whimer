package model

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

const (
	// do not easily change this constant
	WhimerSessId = "WHIMERSESSID"
	WhimerUid    = "uid"
)

const (
	SessionStatusInited  int8 = 0
	SessionStatusActive  int8 = 1
	SessionStatusExpired int8 = 2
	SessionStatusKicked  int8 = 3
)

// session元数据
type SessionMeta struct {
	Id       string `json:"id"`
	ExpireAt int64  `json:"expire_at"` // 过期时间 unix second timestamp
	Status   int8   `json:"status"`    // session 状态
}

func (m *SessionMeta) IsExpired() bool {
	return time.Now().Unix()-1 > m.ExpireAt
}

// 用户session信息
type Session struct {
	Meta     SessionMeta `json:"meta"`
	Uid      int64       `json:"uid"`
	Platform string      `json:"platform"` // session对应平台
	Detail   string      `json:"detail"`   // 用户详细信息
}

func NewSession(id string, expireAt int64) *Session {
	return &Session{
		Meta: SessionMeta{
			Id:       id,
			ExpireAt: expireAt,
		},
	}
}

func (s *Session) Cookie() *http.Cookie {
	return &http.Cookie{
		Name:     WhimerSessId,
		Value:    s.Meta.Id,
		Path:     "/",
		Expires:  time.Unix(s.Meta.ExpireAt, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *Session) UidCookie() *http.Cookie {
	return &http.Cookie{
		Name:    WhimerUid,
		Value:   strconv.FormatInt(s.Uid, 10),
		Path:    "/",
		Expires: time.Unix(s.Meta.ExpireAt, 0),
	}
}

// session序列化
type SessionSerializer interface {
	Serialize(s *Session) ([]byte, error)
	Deserialize([]byte) (*Session, error)
}

type JsonSessionSerializer struct{}

func (s JsonSessionSerializer) Serialize(sess *Session) ([]byte, error) {
	return json.Marshal(sess)
}

func (s JsonSessionSerializer) Deserialize(data []byte) (*Session, error) {
	var sess Session
	err := json.Unmarshal(data, &sess)
	if err != nil {
		return nil, err
	}

	return &sess, nil
}

type MsgpackSessionSerializer struct{}

func (s MsgpackSessionSerializer) Serialize(sess *Session) ([]byte, error) {
	return msgpack.Marshal(sess)
}

func (s MsgpackSessionSerializer) Deserialize(data []byte) (*Session, error) {
	var sess Session
	err := msgpack.Unmarshal(data, &sess)
	if err != nil {
		return nil, err
	}

	return &sess, nil
}
