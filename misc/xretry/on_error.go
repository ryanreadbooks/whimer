package xretry

import "errors"

// 先执行一遍f
//
// 如果f返回target 则重试retryCnt次
func OnError(f func() error, target error, retryCnt int) error {
	err := f()
	if err == nil {
		return nil
	}

	if errors.Is(err, target) {
		for range retryCnt {
			err = f()
			if err == nil {
				return nil
			}
		}
	}

	return err
}
