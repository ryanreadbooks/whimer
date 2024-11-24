package dep

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type ISmsSender interface {
	Send(ctx context.Context, tel string, content string) error
}

type logSmsSender struct{}

func (s *logSmsSender) Send(ctx context.Context, tel string, content string) error {
	xlog.Msg(fmt.Sprintf("Send to %s, content: %s", tel, content)).Infox(ctx)
	return nil
}
