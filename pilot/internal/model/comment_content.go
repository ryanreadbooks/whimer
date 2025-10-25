package model

type CommentContent struct {
	Text    string   `json:"text"`
	AtUsers []AtUser `json:"at_users"`
}
