package relation

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
)

var (
	relationer relationv1.RelationServiceClient
)

func Init(c *config.Config) {
	relationer = xgrpc.NewRecoverableClient(c.Backend.Relation,
		relationv1.NewRelationServiceClient,
		func(cc relationv1.RelationServiceClient) { relationer = cc })
}

func RelationServer() relationv1.RelationServiceClient {
	return relationer
}
