package service

import "github.com/ryanreadbooks/whimer/lambda/media/internal/config"

type Service struct {

}

func New(c *config.Config) *Service {
	return &Service{}
}