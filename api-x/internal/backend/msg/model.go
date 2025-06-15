package msg

type ListChatsReq struct {
	Seq   int64 `form:"seq,optional"`
	Count int   `form:"count,optional"`
}
