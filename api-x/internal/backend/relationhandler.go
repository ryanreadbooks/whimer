package backend

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/relation"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) UserFollowAction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		req, err := xhttp.ParseValidate[relation.FollowReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 关注或者取消关注
		resp, err := relation.RelationServer().FollowUser(ctx, &relationv1.FollowUserRequest{
			Follower: uid,
			Followee: req.Target,
			Action:   relationv1.FollowUserRequest_Action(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_ = resp
		xhttp.OkJson(w, nil)
	}
}
