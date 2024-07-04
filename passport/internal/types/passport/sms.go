package passport

type SmsSendReq struct {
	Tel  string `json:"tel"`          // 手机号
	Zone string `json:"cid,optional"` // TODO 手机区号
	// TODO 补充验证码相关结果字段
}

type SmdSendRes struct {
}
