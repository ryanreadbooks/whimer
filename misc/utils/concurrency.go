package utils

import (
	"github.com/zeromicro/go-zero/core/logx"
)

func SafeGo(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logx.ErrorStackf("panic: %v", err)
			}
		}()

		f()
	}()
}
