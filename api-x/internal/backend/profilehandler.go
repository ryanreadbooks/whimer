package backend

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/note"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/relation"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
	"github.com/zeromicro/go-zero/rest/httpx"

	relationv1 "github.com/ryanreadbooks/whimer/relation/sdk/v1"

	"golang.org/x/sync/errgroup"
)

// 获取用户的投稿数量、点赞数量等信息
func (h *Handler) GetProfileStat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			uid  = metadata.Uid(r.Context())
			stat struct {
				Posted     uint64 `json:"posted"`
				Fans       uint64 `json:"fans"`
				Followings uint64 `json:"followings"`
			}
		)

		eg, ctx := errgroup.WithContext(r.Context())

		// 用户投稿数量
		eg.Go(func() error {
			resp, err := note.NoteCreatorServer().GetPostedCount(ctx, &notev1.GetPostedCountRequest{
				Uid: uid,
			})
			if err != nil {
				return err
			}

			stat.Posted = resp.Count
			return nil
		})

		// 用户粉丝数量
		eg.Go(func() error {
			resp, err := relation.RelationServer().GetUserFanCount(ctx,
				&relationv1.GetUserFanCountRequest{
					Uid: uid,
				})
			if err != nil {
				return err
			}

			stat.Fans = resp.Count
			return nil
		})

		eg.Go(func() error {
			// 用户关注数量
			resp, err := relation.RelationServer().GetUserFollowingCount(ctx,
				&relationv1.GetUserFollowingCountRequest{
					Uid: uid,
				})
			if err != nil {
				return err
			}

			stat.Followings = resp.Count
			return nil
		})

		err := eg.Wait()
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &stat)
	}
}
