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
	relationer = relationv1.NewRelationServiceClient(
		xgrpc.NewRecoverableClientConn(c.Backend.Relation),
	)
}

func RelationServer() relationv1.RelationServiceClient {
	return relationer
}
