package procedure

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

var ErrProcedureNotRegistered = fmt.Errorf("procedure type not registered")

// PollState 轮询结果状态
type PollState int

const (
	PollStateRunning PollState = iota // 运行中
	PollStateSuccess                  // 成功
	PollStateFailure                  // 失败
)

type ProcedureResult struct {
	NoteId int64
	TaskId string
	Arg    any
}

// Procedure 流程处理器接口
// 每种流程类型（资源处理、审核等）需要实现该接口
//
// 下列方法中的note参数为基础信息 不包含额外信息 比如ext/at_user/assets等 需要自行获取
type Procedure interface {
	// Type 返回流程类型
	Type() model.ProcedureType

	// 流程开始前的初始化工作 一般会作为本地事务的一部分
	//
	// 返回doRecord=true表示需要该流程接下来会进行远程调用 需要创建本地调用记录
	BeforeExecute(ctx context.Context, note *model.Note) (doRecord bool, err error)

	// Execute 执行流程任务 返回任务ID用于后续追踪
	//
	// 一般是执行一次远程调用 返回任务ID用于后续追踪
	Execute(ctx context.Context, note *model.Note) (taskId string, err error)

	// 中断执行任务
	ObAbort(ctx context.Context, note *model.Note, taskId string) error

	// OnSuccess 流程成功处理 更新笔记状态、记录状态等
	//
	// 会在本地事务中执行
	//
	// 返回true表示需要更新记录状态
	OnSuccess(ctx context.Context, result *ProcedureResult) (bool, error)

	// OnFailure 流程失败处理 更新笔记状态、记录状态等
	//
	// 会在本地事务中执行
	//
	// 返回true表示需要更新记录状态
	//
	// 当因为重试次数满导致失败时传入的Arg为nil
	OnFailure(ctx context.Context, result *ProcedureResult) (bool, error)

	// PollResult 主动轮询任务结果
	//
	// 用于后台重试时检查已提交任务的状态
	//
	// 返回: PollStateSuccess/PollStateFailure/PollStateRunning
	//
	// 返回PollStateSuccess时会触发OnSuccess;
	// 返回PollStateFailure时会触发OnFailure;
	// 返回PollStateRunning时等待下一次轮询
	PollResult(ctx context.Context, record *biz.ProcedureRecord) (PollState, any, error)

	// Retry 重试流程
	//
	// record 包含当前重试状态信息
	//
	// 返回重试的taskId
	Retry(ctx context.Context, record *biz.ProcedureRecord) (string, error)
}

// 实现此接口的流程会在本地流程完成后自动标记完成
type AutoCompleter interface {
	// 自动完成
	//
	// 有些流程没有回调触发OnSuccess 需要手动调用完成,
	// 如果设置了自动成功就会自动触发OnSuccess或者OnFailure方法
	//
	// success: true=OnSuccess, false=OnFailure
	// autoComplete: true=自动完成, false=不需要自动完成
	AutoComplete(ctx context.Context, note *model.Note, taskId string) (success, autoComplete bool, arg any)
}

// 实现此接口以提供创建本地流程的参数
type ProcedureParamProvider interface {
	// 提供参数 该参数会被保存以便后续重试时使用
	//
	// 参数的序列化和反序列化由各流程自行负责
	Provide(note *model.Note) []byte
}
