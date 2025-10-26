package model

type NotifyMsgLocation string

const (
	NotifyMsgOnNote    NotifyMsgLocation = "on_note"    // 对笔记的操作
	NotifyMsgOnComment NotifyMsgLocation = "on_comment" // 对评论的操作
)

type MsgStatus string

const (
	MsgStatusNormal   = "normal"
	MsgStatusNoReveal = "noreveal"
	MsgStatusRevoked  = "revoked"
)
