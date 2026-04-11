package dep

import (
	relationv1 "github.com/ryanreadbooks/whimer/idl/gen/go/relation/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
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
