package errors

import (
	stderr "errors"

	"github.com/ryanreadbooks/whimer/misc/xerror"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

func Convert(err error) error {
	if err == nil {
		return nil
	}

	var e *types.ElasticsearchError
	if stderr.As(err, &e) {
		return xerror.Wrapf(xerror.ErrInternal, "%s", *e.ErrorCause.Reason)
	}

	return xerror.Wrapf(xerror.ErrInternal, "%s", err.Error())
}
