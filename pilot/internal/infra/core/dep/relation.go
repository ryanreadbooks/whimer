package dep

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
)

var (
	relationer relationv1.RelationServiceClient
)

func InitRelation(c *config.Config) {
	relationer = xgrpc.NewRecoverableClient(c.Backend.Relation,
		relationv1.NewRelationServiceClient,
		func(cc relationv1.RelationServiceClient) { relationer = cc })
}

func RelationServer() relationv1.RelationServiceClient {
	return relationer
}
