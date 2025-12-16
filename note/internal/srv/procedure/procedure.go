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

// PollState 轮询结果状态
type PollState int

const (
	PollStateRunning PollState = iota // 运行中
	PollStateSuccess                  // 成功
	PollStateFailure                  // 失败
)

// Procedure 流程处理器接口
// 每种流程类型（资源处理、审核等）需要实现该接口
type Procedure interface {
	// Type 返回流程类型
	Type() model.ProcedureType

	// 流程开始前的初始化工作
	//
	// 返回doRecord=true表示需要该流程接下来会进行远程调用 需要创建本地调用记录
	//
	// 返回doRecord=false后, 后续的Execute失败时将不会触发重试机制
	PreStart(ctx context.Context, note *model.Note) (doRecord bool, err error)

	// Execute 执行流程任务 返回任务ID用于后续追踪
	Execute(ctx context.Context, note *model.Note) (taskId string, err error)

	// OnSuccess 流程成功处理 更新笔记状态、记录状态等
	//
	// 会在本地事务中执行
	//
	// 返回true表示需要更新记录状态
	OnSuccess(ctx context.Context, noteId int64, taskId string) (bool, error)

	// OnFailure 流程失败处理 更新笔记状态、记录状态等
	//
	// 会在本地事务中执行
	//
	// 返回true表示需要更新记录状态
	OnFailure(ctx context.Context, noteId int64, taskId string) (bool, error)

	// PollResult 主动轮询任务结果
	//
	// 用于后台重试时检查已提交任务的状态
	//
	// 返回: PollStateSuccess/PollStateFailure/PollStateRunning
	PollResult(ctx context.Context, taskId string) (PollState, error)

	// Retry 重试流程
	//
	// record 包含当前重试状态信息
	Retry(ctx context.Context, record *biz.ProcedureRecord) error
}

type AutoCompleter interface {
	// 自动完成
	//
	// 有些流程没有回调触发OnSuccess 需要手动调用完成
	//
	// success: true=OnSuccess, false=OnFailure
	// autoComplete: true=自动完成, false=不需要自动完成
	AutoComplete(ctx context.Context, note *model.Note, taskId string) (success, autoComplete bool)
}
