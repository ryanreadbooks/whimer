package xerror

import (
	"context"
	"testing"
)

func TestHander_ErrHandler(t *testing.T) {
	// service call
	dao := func() error {
		return Propagate(ErrInternal.Msg("dao出错")).WithField("sql", "select id from test")
	}

	service := func() error {
		err := dao()
		if err != nil {
			return PropagateMsg(err, "service failed").
				WithCtx(context.Background()).
				WithField("service", "test").
				WithExtra("name", "test")
		}

		return nil
	}

	entry := func() error {
		err := service()
		if err != nil {
			return PropagateMsg(err, "entry failed").WithField("path", "/api/v1/entry")
		}
		return nil
	}

	err := entry()
	errorHandler(err)
}
