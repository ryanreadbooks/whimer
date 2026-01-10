package worker

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"syscall"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsRetryableError 判断错误是否可重试
// 可重试的错误通常是临时性的，如超时、网络错误、资源暂时不可用等
// 不可重试的错误通常是永久性的，如参数错误、资源不存在等
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// context 相关错误
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return false // 取消通常是主动行为，不应重试
	}

	// IO 错误
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, io.EOF) {
		return true
	}

	// 网络错误
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	// DNS 错误（临时性）
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.Temporary()
	}

	// 系统调用错误
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	if errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	if errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}
	if errors.Is(err, syscall.ENOBUFS) {
		return true
	}

	// 临时文件系统错误
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	// gRPC 错误
	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.DeadlineExceeded,
			codes.ResourceExhausted,
			codes.Unavailable,
			codes.Aborted:
			return true
		case codes.InvalidArgument,
			codes.NotFound,
			codes.AlreadyExists,
			codes.PermissionDenied,
			codes.FailedPrecondition,
			codes.Unimplemented,
			codes.Unauthenticated:
			return false
		}
	}

	return false
}

// RetryableResult 创建一个可重试的失败结果
func RetryableResult(err error) Result {
	return Result{
		Error:     err,
		Retryable: true,
	}
}

// NonRetryableResult 创建一个不可重试的失败结果
func NonRetryableResult(err error) Result {
	return Result{
		Error:     err,
		Retryable: false,
	}
}

// AutoRetryResult 根据错误类型自动判断是否可重试
func AutoRetryResult(err error) Result {
	return Result{
		Error:     err,
		Retryable: IsRetryableError(err),
	}
}

// SuccessResult 创建一个成功的结果
func SuccessResult(output any) Result {
	return Result{
		Output: output,
	}
}
