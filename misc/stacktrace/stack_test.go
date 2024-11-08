package stacktrace

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestStdCaller(t *testing.T) {
	api()
}

func TestStack(t *testing.T) {
	apiStack()
}

func TestCaller(t *testing.T) {
	apiCall()
}

func dao() error {
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	if n == 0 {
		return nil
	}

	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)

	var sbd strings.Builder
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&sbd, "%s()\n\t%s:%d %#x\n", frame.Function, frame.File, frame.Line, frame.Entry)
		if !more {
			break
		}
	}

	fmt.Println(sbd.String())
	return nil
}

func service() error {
	return dao()
}

func api() error {
	return service()
}

func daoStack() error {
	var buf = make([]byte, 1024)
	n := runtime.Stack(buf, false)
	fmt.Printf("%s\n", buf[:n])
	return nil
}

func apiStack() error {
	return daoStack()
}

func daoCaller() {
	fs := TraceStack()
	fmt.Println(FormatFrames(fs))
}

func apiCall() {
	daoCaller()
}