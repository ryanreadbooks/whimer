package model

type ChatType int8

const (
	P2PChat   ChatType = 1
	GroupChat ChatType = 2
)

type ChatStatus int8

const (
	ChatStatusNormal ChatStatus = 0
)
