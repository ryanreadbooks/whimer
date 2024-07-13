package profile

import (
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/misc/oss/uploader"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/repo"
)

type Service struct {
	c              *config.Config
	repo           *repo.Repo
	avatarKeyGen   *keygen.Generator
	avatarUploader *uploader.Uploader
}

func New(c *config.Config, repo *repo.Repo) *Service {
	s := &Service{
		c:    c,
		repo: repo,
		avatarKeyGen: keygen.NewGenerator(
			keygen.WithBucket(c.Oss.Bucket),
			keygen.WithPrefix(c.Oss.Prefix),
			keygen.WithPrependPrefix(true),
			keygen.WithPrependBucket(false), // 生成key时不需要附带上bucket
		),
	}

	avatartUploader, err := uploader.New(uploader.Config{
		Ak:       c.Oss.Ak,
		Sk:       c.Oss.Sk,
		Endpoint: c.Oss.Endpoint,
		Location: c.Oss.Location,
	})
	if err != nil {
		panic(err)
	}
	s.avatarUploader = avatartUploader

	return s
}
