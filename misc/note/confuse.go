package note

import "github.com/ryanreadbooks/whimer/misc/safety"

const (
	noteIdConfuserSalt = "0x7c00:noteIdConfuser:.$35%io"
)

func NewConfuser() *safety.Confuser {
	return safety.NewConfuser(noteIdConfuserSalt, 24)
}