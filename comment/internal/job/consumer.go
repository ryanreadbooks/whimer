package job

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/queue"
	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type Job struct {
	Svc *svc.ServiceContext
}

func New(svc *svc.ServiceContext) *Job {
	j := &Job{
		Svc: svc,
	}

	return j
}

// Job需要实现kq.ConsumerHandler接口
func (j *Job) Consume(ctx context.Context, key, value string) error {
	logx.Infof("job Consume key: %s, value: %s", key, value)
	var data queue.Data
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		logx.Errorf("job consumer json.Unmarshal err: %v", err)
		return err
	}

	switch data.Action {
	case queue.ActAddReply:
		return j.Svc.CommentSvc.BizAddReply(ctx, data.AddReplyData)
	case queue.ActDelReply:
		return j.Svc.CommentSvc.BizDelReply(ctx, data.DelReplyData)
	default:
		logx.Errorf("job consumer got unsupported action type: %v", data.Action)
		return global.ErrInternal
	}
}
