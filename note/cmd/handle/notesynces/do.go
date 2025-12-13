package notesynces

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/data"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/note/internal/model/convert"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
	"github.com/ryanreadbooks/whimer/note/pkg/id"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

func Handle(c *config.Config, bizz *biz.Biz, svc *srv.Service, dt *data.Data) error {
	var (
		cursor    int64 = math.MaxInt64
		batchsize int32 = 500
		ctx             = context.Background()
	)

	id.InitNoteIdObfuscate(c.Obfuscate.Note.Options()...)
	id.InitTagIdObfuscate(c.Obfuscate.Tag.Options()...)

	var userMaps = make(map[int64]string)

	for {
		batchNotes, err := dt.Note.ListPublicByCursor(ctx, cursor, batchsize)
		if err != nil {
			return fmt.Errorf("dao list public by cursor err: %w", err)
		}

		if len(batchNotes) == 0 || errors.Is(err, xsql.ErrNoRecord) {
			break
		}

		xlog.Msgf("list note by cursor got %d notes", len(batchNotes)).Info()

		res, err := bizz.Note.AssembleNotes(ctx, convert.NoteSliceFromDao(batchNotes))
		if err != nil {
			return fmt.Errorf("creator assemble notes err: %w", err)
		}

		err = bizz.Note.AssembleNotesExt(ctx, res.Items)
		if err != nil {
			return fmt.Errorf("creator assemble note exts err: %w", err)
		}

		reqs := make([]*searchv1.Note, 0)

		missingUids := []int64{}
		for _, item := range res.Items {
			if _, ok := userMaps[item.Owner]; !ok {
				missingUids = append(missingUids, item.Owner)
			}
		}

		if len(missingUids) > 0 {
			userResp, err := dep.GetUserer().BatchGetUser(ctx, &userv1.BatchGetUserRequest{
				Uids: missingUids,
			})
			if err != nil {
				return fmt.Errorf("batch get user err: %w", err)
			}

			for _, u := range userResp.Users {
				userMaps[u.Uid] = u.Nickname
			}
		}

		for _, item := range res.Items {
			noteId := id.NoteId(item.NoteId).String()
			tagList := make([]*searchv1.NoteTag, 0, len(item.Tags))
			for _, tag := range item.Tags {
				tagId := id.TagId(tag.Id).String()
				tagList = append(tagList, &searchv1.NoteTag{
					Id:    string(tagId),
					Name:  tag.Name,
					Ctime: tag.Ctime,
				})
			}

			assetType := searchv1.Note_ASSET_TYPE_IMAGE // for now
			reqs = append(reqs, &searchv1.Note{
				NoteId:   noteId,
				Title:    item.Title,
				Desc:     item.Desc,
				CreateAt: item.CreateAt,
				UpdateAt: item.UpdateAt,
				Author: &searchv1.Note_Author{
					Uid:      item.Owner,
					Nickname: userMaps[item.Owner],
				},
				TagList:       tagList,
				Visibility:    searchv1.Note_VISIBILITY_PUBLIC,
				AssetType:     assetType,
				LikesCount:    item.Likes,
				CommentsCount: item.Replies,
			})
		}

		_, errDoc := dep.GetSearchDocer().BatchAddNote(ctx,
			&searchv1.BatchAddNoteRequest{
				Notes: reqs,
			})
		if errDoc != nil {
			xlog.Msg("search docer batch add note failed").Err(errDoc).Error()
		}

		cursor = batchNotes[len(batchNotes)-1].Id
	}

	return nil
}
