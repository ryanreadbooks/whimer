package procedure

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

var (
	ErrProcedureNotRegistered = fmt.Errorf("procedure type not registered")
)

// Procedure 流程处理器接口
// 每种流程类型（资源处理、审核等）需要实现该接口
type Procedure interface {
	// Type 返回流程类型
	Type() model.ProcedureType

	// 流程开始前的初始化工作
	PreStart(ctx context.Context, note *model.Note) error

	// Execute 执行流程任务 返回任务ID用于后续追踪
	Execute(ctx context.Context, note *model.Note) (taskId string, err error)

	// OnSuccess 流程成功处理 更新笔记状态、记录状态等
	OnSuccess(ctx context.Context, noteId int64, taskId string) error

	// OnFailure 流程失败处理 更新笔记状态、记录状态等
	OnFailure(ctx context.Context, noteId int64, taskId string) error

	// PollResult 主动轮询任务结果
	// 用于后台重试时检查已提交任务的状态
	// 返回: success=true表示任务成功, success=false表示任务失败
	PollResult(ctx context.Context, taskId string) (success bool, err error)

	// Retry 重试流程
	// record 包含当前重试状态信息
	Retry(ctx context.Context, record *biz.ProcedureRecord) error
}
