package convert

import (
	relationv1 "github.com/ryanreadbooks/whimer/idl/gen/go/relation/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/vo"
)

func VoFollowActionToPb(action vo.FollowAction) relationv1.FollowUserRequest_Action {
	switch action {
	case vo.ActionFollow:
		return relationv1.FollowUserRequest_ACTION_FOLLOW
	case vo.ActionUnFollow:
		return relationv1.FollowUserRequest_ACTION_UNFOLLOW
	}
	return relationv1.FollowUserRequest_ACTION_UNSPECIFIED
}
