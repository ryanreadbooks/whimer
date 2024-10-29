package xerror

import (
	"fmt"
	"runtime"

	"github.com/ryanreadbooks/whimer/misc/stacktrace"
)

type Unwrapper interface {
	Unwrap() error
}

type Causer interface {
	Cause() error
}

var emptyFrames stacktrace.Frames

type errStack struct {
	cause error
	msg   string
	stack []*runtime.Frame
}

func (e *errStack) Error() string {
	if e.msg != "" {
		return e.cause.Error() + "(" + e.msg + ")"
	}
	return e.cause.Error()
}

func (e *errStack) Format(f fmt.State, verb rune) {
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

func (e *errStack) Unwrap() error {
	return e.cause
}

func (e *errStack) Cause() error {
	return e.cause
}

type Stacker interface{ Stack() []*runtime.Frame }

func (e *errStack) Stack() []*runtime.Frame {
	return e.stack
}

func Propagate(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(Stacker); ok {
		return &errStack{
			cause: err,
		}
	}

	return &errStack{
		cause: err,
		stack: stacktrace.TraceStack(),
	}
}

func PropagateMsg(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(Stacker); ok {
		return &errStack{
			cause: err,
			msg:   fmt.Sprintf(format, args...),
		}
	}

	return &errStack{
		cause: err,
		msg:   fmt.Sprintf(format, args...),
		stack: stacktrace.TraceStack(),
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

func HasFramesHold(err error) bool {
	return len(UnwindFrames(err)) != 0
}

func UnwindFrames(err error) stacktrace.Frames {
	var frames stacktrace.Frames
	for err != nil {
		stacker, ok := err.(Stacker)
		if ok {
			if stack := stacker.Stack(); len(stack) != 0 {
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
