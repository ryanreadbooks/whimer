package dto

import (
	notifyentity "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/entity"
	uservo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
)

type MentionMsgWithUser struct {
	*notifyentity.MentionedMsg
	User *uservo.User `json:"user,omitempty"`
}

type ReplyMsgWithUser struct {
	*notifyentity.ReplyMsg
	User *uservo.User `json:"user,omitempty"`
}

type LikesMsgWithUser struct {
	*notifyentity.LikesMsg
	User *uservo.User `json:"user,omitempty"`
}

type ListUserMentionMsgResult struct {
	Msgs    []*MentionMsgWithUser
	ChatId  string
	HasNext bool
}

type ListUserReplyMsgResult struct {
	Msgs    []*ReplyMsgWithUser
	ChatId  string
	HasNext bool
}

type ListUserLikesMsgResult struct {
	Msgs    []*LikesMsgWithUser
	ChatId  string
	HasNext bool
}
