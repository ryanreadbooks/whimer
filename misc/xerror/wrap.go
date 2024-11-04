package xerror

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/stacktrace"
)

type ErrProxy interface {
	error
	Stack() []*runtime.Frame
	Context() context.Context
	Fields() map[string]any
	Extra() map[string]any
	WithCtx(ctx context.Context) ErrProxy
	WithField(key string, val any) ErrProxy
	WithExtra(key string, val any) ErrProxy
	WithFields(kvs ...any) ErrProxy
	WithExtras(kvs ...any) ErrProxy
}

type Unwrapper interface {
	Unwrap() error
}

type Causer interface {
	Cause() error
}

var emptyFrames stacktrace.Frames

type errProxy struct {
	cause error
	msg   string
	stack []*runtime.Frame

	// log related fields
	ctx    context.Context
	fields map[string]any
	extra  map[string]any
}

func (e *errProxy) Error() string {
	// if e.msg != "" {
	// 	return e.cause.Error() + "(" + e.msg + ")"
	// }
	// return e.cause.Error()
	return e.msg
}

func (e *errProxy) Format(f fmt.State, verb rune) {
	if e == nil {
		return
	}

	switch verb {
	case 'v':
		if len(e.msg) != 0 {
			fmt.Fprintf(f, "%s <- ", e.msg)
		}
		fmt.Fprintf(f, "%v", e.cause)
		if len(e.stack) != 0 {
			fmt.Fprintf(f, "\n\n%s", stacktrace.FormatFrames(e.stack))
		}
	default:
		fmt.Fprintf(f, "%s", e.Error())
	}
}

func (e *errProxy) Unwrap() error {
	return e.cause
}

func (e *errProxy) Cause() error {
	return e.cause
}

func (e *errProxy) Stack() []*runtime.Frame {
	return e.stack
}

func (e *errProxy) Context() context.Context {
	return e.ctx
}

func (e *errProxy) Fields() map[string]any {
	return e.fields
}

func (e *errProxy) Extra() map[string]any {
	return e.extra
}

// log related methods
func (e *errProxy) WithCtx(ctx context.Context) ErrProxy {
	e.ctx = ctx
	return e
}

func (e *errProxy) WithField(key string, val any) ErrProxy {
	e.fields[key] = val
	return e
}

func (e *errProxy) WithExtra(key string, val any) ErrProxy {
	e.extra[key] = val
	return e
}

func (e *errProxy) WithFields(kvs ...any) ErrProxy {
	if len(kvs)%2 == 0 {
		for i := range kvs {
			e.fields[fmt.Sprintf("%v", kvs[i*2])] = kvs[i*2+1]
		}
	}

	return e
}

func (e *errProxy) WithExtras(kvs ...any) ErrProxy {
	if len(kvs)%2 == 0 {
		for i := range kvs {
			e.extra[fmt.Sprintf("%v", kvs[i*2])] = kvs[i*2+1]
		}
	}

	return e
}

func Wrap(err error) ErrProxy {
	if err == nil {
		return nil
	}

	if pxy, ok := err.(ErrProxy); ok {
		// combine with previous errProxy
		return &errProxy{
			cause:  err,
			fields: pxy.Fields(),
			extra:  pxy.Extra(),
			ctx:    pxy.Context(),
		}
	}

	return &errProxy{
		cause:  err,
		stack:  stacktrace.TraceStack(),
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
}

func Wrapf(err error, format string, args ...any) ErrProxy {
	if err == nil {
		return nil
	}

	if pxy, ok := err.(ErrProxy); ok {
		return &errProxy{
			cause:  err,
			msg:    fmt.Sprintf(format, args...),
			fields: pxy.Fields(),
			extra:  pxy.Extra(),
			ctx:    pxy.Context(),
		}
	}

	return &errProxy{
		cause:  err,
		msg:    fmt.Sprintf(format, args...),
		stack:  stacktrace.TraceStack(),
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
}

func Cause(err error) error {
	for err != nil {
		causer, ok := err.(interface{ Cause() error })
		if !ok {
			break
		}
		err = causer.Cause()
	}

	return err
}

func FramesWrapped(err error) bool {
	return len(UnwrapFrames(err)) != 0
}

func UnwrapFrames(err error) stacktrace.Frames {
	var frames stacktrace.Frames
	for err != nil {
		proxyer, ok := err.(ErrProxy)
		if ok {
			if stack := proxyer.Stack(); len(stack) != 0 {
				frames = stack
			}
		}

		causer, ok := err.(Causer)
		if !ok {
			break
		}
		err = causer.Cause()
	}

	return frames
}

func UnwrapMsg(err error) (string, error) {
	var (
		msbd          strings.Builder
		underlyingErr error
	)

	for err != nil {
		var msg string
		underlyingErr = err
		if msg = err.Error(); len(msg) > 0 {
			msbd.WriteString(msg)
		}

		causer, ok := err.(Causer)
		if !ok {
			break
		}
		err = causer.Cause()
		if len(msg) > 0 && err != nil {
			msbd.WriteString(" -> ")
		}
	}

	return msbd.String(), underlyingErr
}
