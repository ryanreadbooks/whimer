package producer

import (
	"context"
	"encoding/json"
	"time"

	taskv1 "github.com/ryanreadbooks/whimer/conductor/api/task/v1"
	taskservice "github.com/ryanreadbooks/whimer/conductor/api/taskservice/v1"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ClientOptions 客户端配置
type ClientOptions struct {
	HostConf zrpc.RpcClientConf
	// 默认命名空间
	Namespace string
}

// Client 任务客户端
type Client struct {
	opts   ClientOptions
	client taskservice.TaskServiceClient
}

// NewClient 创建任务客户端
func NewClient(opts ClientOptions) (*Client, error) {
	cli, err := zrpc.NewClient(opts.HostConf)
	if err != nil {
		return nil, err
	}

	return &Client{
		opts:   opts,
		client: taskservice.NewTaskServiceClient(cli.Conn()),
	}, nil
}

// MustNewClient 创建任务客户端，失败则 panic
func MustNewClient(opts ClientOptions) *Client {
	cli, err := NewClient(opts)
	if err != nil {
		panic(err)
	}
	return cli
}

// Task 任务信息
type Task struct {
	Id          string
	Namespace   string
	TaskType    string
	InputArgs   []byte
	OutputArgs  []byte
	CallbackUrl string
	State       string
	MaxRetryCnt int64
	ExpireTime  int64
	Ctime       int64
	Utime       int64
	TraceId     string
}

// UnmarshalOutput 反序列化输出结果
func (t *Task) UnmarshalOutput(v any) error {
	if len(t.OutputArgs) == 0 {
		return nil
	}
	return json.Unmarshal(t.OutputArgs, v)
}

func taskFromProto(t *taskv1.Task) *Task {
	if t == nil {
		return nil
	}
	return &Task{
		Id:          t.Id,
		Namespace:   t.Namespace,
		TaskType:    t.TaskType,
		InputArgs:   t.InputArgs,
		OutputArgs:  t.OutputArgs,
		CallbackUrl: t.CallbackUrl,
		State:       t.State,
		MaxRetryCnt: t.MaxRetryCnt,
		ExpireTime:  t.ExpireTime,
		Ctime:       t.Ctime,
		Utime:       t.Utime,
		TraceId:     t.TraceId,
	}
}

// ExecuteOptions 执行任务的选项
type ExecuteOptions struct {
	// 命名空间（可选，不设置则使用 ClientOptions 中的默认值）
	Namespace string

	// 执行成功的回调 URL
	//
	// 失败 过期等状态均不回调
	CallbackUrl string

	// 最大重试次数 (-1: 无限重试, 0: 不重试, >0: 指定次数)
	MaxRetry int64

	// 过期时间 精度到秒
	ExpireTime time.Time

	// 过期时长（与 ExpireTime 二选一） 精度到秒
	ExpireAfter time.Duration
}

// Execute 执行任务
// taskType: 任务类型
// input: 输入参数（自动 JSON 序列化，传 nil 表示无参数）
// opts: 执行选项
func (c *Client) Execute(ctx context.Context, taskType string, input any, opts ExecuteOptions) (string, error) {
	var inputArgs []byte
	var err error
	if input != nil {
		inputArgs, err = json.Marshal(input)
		if err != nil {
			return "", err
		}
	}

	return c.ExecuteRaw(ctx, taskType, inputArgs, opts)
}

// ExecuteRaw 执行任务（原始字节输入）
func (c *Client) ExecuteRaw(ctx context.Context, taskType string, inputArgs []byte, opts ExecuteOptions) (string, error) {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = c.opts.Namespace
	}

	req := &taskservice.RegisterTaskRequest{
		TaskType:    taskType,
		Namespace:   namespace,
		InputArgs:   inputArgs,
		CallbackUrl: opts.CallbackUrl,
		MaxRetryCnt: opts.MaxRetry,
	}

	if !opts.ExpireTime.IsZero() {
		req.ExpireTime = timestamppb.New(opts.ExpireTime)
	} else if opts.ExpireAfter > 0 {
		req.ExpireTime = timestamppb.New(time.Now().Add(opts.ExpireAfter))
	}

	resp, err := c.client.RegisterTask(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.TaskId, nil
}

// GetTask 获取任务信息
func (c *Client) GetTask(ctx context.Context, taskId string) (*Task, error) {
	resp, err := c.client.GetTask(ctx, &taskservice.GetTaskRequest{
		TaskId: taskId,
	})
	if err != nil {
		return nil, err
	}

	return taskFromProto(resp.Task), nil
}

// AbortTask 终止任务
func (c *Client) AbortTask(ctx context.Context, taskId string) error {
	_, err := c.client.AbortTask(ctx, &taskservice.AbortTaskRequest{
		TaskId: taskId,
	})
	return err
}
