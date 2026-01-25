package relation

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/repository"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
)

type RelationAdapterImpl struct {
	relationCli relationv1.RelationServiceClient
}

var _ repository.RelationAdapter = &RelationAdapterImpl{}

func NewRelationAdapterImpl(c relationv1.RelationServiceClient) *RelationAdapterImpl {
	return &RelationAdapterImpl{
		relationCli: c,
	}
}

func (a *RelationAdapterImpl) BatchGetFollowStatus(
	ctx context.Context, uid int64, targets []int64,
) (map[int64]bool, error) {
	resp, err := a.relationCli.BatchCheckUserFollowed(ctx,
		&relationv1.BatchCheckUserFollowedRequest{
			Uid:     uid,
			Targets: targets,
		})
	if err != nil {
		return nil, err
	}
	
	return resp.Status, nil
}
