package signinup

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/repo"

	"github.com/ryanreadbooks/folium/sdk"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Service struct {
	c     *config.Config
	repo  *repo.Repo
	cache *redis.Redis
	idgen *sdk.Client
}

func New(c *config.Config, repo *repo.Repo, cache *redis.Redis) *Service {
	s := &Service{
		c:     c,
		repo:  repo,
		cache: cache,
	}

	var err error
	s.idgen, err = sdk.NewClient(sdk.WithGrpc(s.c.Idgen.Addr))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = s.idgen.Ping(ctx)
	if err != nil {
		logx.Errorf("new passport svc, can not ping idgen(folium): %v", err)
	}

	return s
}
