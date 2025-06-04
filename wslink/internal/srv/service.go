package srv

import (
	"context"
	"fmt"
	"strings"

	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
)

type Service struct {
	Config *config.Config
}

func NewService(c *config.Config) *Service {
	return &Service{
		Config: c,
	}
}

func (s *Service) OnCreate(ctx context.Context, sess *ws.Session) error {
	println(sess.GetId())
	return nil
}

func (s *Service) OnData(ctx context.Context, sess *ws.Session, data []byte) error {
	fmt.Printf("data reached on %s, %s\n", sess.GetId(), data)
	// do echo
	return sess.WriteText(strings.ToUpper(string(data)))
}

func (s *Service) OnClosed(ctx context.Context, sess *ws.Session) error {
	fmt.Printf("sess %s closed\n", sess.GetId())
	return nil
}
