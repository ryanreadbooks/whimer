package stacktrace

import (
	"fmt"
	"runtime"
	"strings"
)

type Frames []*runtime.Frame

func (fs Frames) FormatLines() []string {
	var lines []string
	for _, f := range fs {
		lines = append(lines, fmt.Sprintf("%s:%d", f.File, f.Line))
	}
	return lines
}

func (fs Frames) FormatFuncs() []string {
	var funcs []string
	for _, f := range fs {
		funcs = append(funcs, f.Function)
	}
	return funcs
}

func TraceStack() Frames {
	var pcs = make([]uintptr, 1024)
	n := runtime.Callers(3, pcs)
	if n == 0 {
		return Frames{}
	}

	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)
	res := make([]*runtime.Frame, 0, n)
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		res = append(res, &frame)
	}

	return res
}

func FormatFrames(frames Frames) string {
	var sbd strings.Builder
	sbd.Grow(512)
	for _, frame := range frames {
		// fmt.Fprintf(&sbd, "%s()\n\t%s:%d %#x\n", frame.Function, frame.File, frame.Line, frame.Entry)
		fmt.Fprintf(&sbd, "%s()\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
	}

	return strings.TrimSpace(sbd.String())
}
