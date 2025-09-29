package middleware

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/stacktrace"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recover
			if e := recover(); e != nil {
				logErr := xerror.Wrapf(xerror.ErrPanic, "%v", e)
				xlog.Msg("panic").Err(logErr).Extra("stack", stacktrace.FormatFrames(xerror.UnwrapFrames(logErr))).Error()

				// we still need to response to the client
				httpx.Error(w, xerror.ErrInternal)
			}
		}()

		next(w, r)
	}
}
