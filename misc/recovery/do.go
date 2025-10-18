package recovery

import (
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/stacktrace"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

func Do(fn func() error) (err error) {
	defer func() {
		// recover
		if e := recover(); e != nil {
			logErr := xerror.Wrapf(xerror.ErrInternalPanic, fmt.Sprintf("%v", e))
			xlog.Msg("panic").Err(logErr).Extra("stack", stacktrace.FormatFrames(xerror.UnwrapFrames(logErr))).Error()

			err = logErr
		}
	}()

	return fn()
}

func DoV2(fn func() error) func() error {
	return func() error {
		return Do(fn)
	}
}
