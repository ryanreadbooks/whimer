package sysmsg

type DeletionEvent struct {
	MsgId string `json:"msg_id"` // 待删除消息id
	Uid   int64  `json:"uid"`    // 消息所属用户
}
