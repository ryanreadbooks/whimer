package model

type SessionStatus int8

const (
	SessionStatusNoActive SessionStatus = 0
	SessionStatusActive   SessionStatus = 1
)
