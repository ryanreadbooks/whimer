package note

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

// IsNotePrivate 判断笔记是否私密
func IsNotePrivate(note *notev1.NoteItem) bool {
	return note.GetPrivacy() == int32(notev1.NotePrivacy_PRIVATE)
}

// AsyncNoteToSearcher 同步笔记到搜索引擎
func (b *Biz) AsyncNoteToSearcher(ctx context.Context, noteId int64, note *notev1.NoteItem) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: fmt.Sprintf("note_creator_sync_note_%d", noteId),
		Job: func(ctx context.Context) error {
			if IsNotePrivate(note) {
				return nil
			}

			nid := imodel.NoteId(noteId).String()
			tagList := make([]*searchv1.NoteTag, 0, len(note.GetTags()))
			for _, tag := range note.GetTags() {
				tagId := imodel.TagId(tag.GetId()).String()
				tagList = append(tagList, &searchv1.NoteTag{
					Id:    string(tagId),
					Name:  tag.GetName(),
					Ctime: tag.GetCtime(),
				})
			}

			vis := searchv1.Note_VISIBILITY_PUBLIC
			if IsNotePrivate(note) {
				vis = searchv1.Note_VISIBILITY_PRIVATE
			}
			assetType := searchv1.Note_ASSET_TYPE_IMAGE

			docNote := []*searchv1.Note{{
				NoteId:   string(nid),
				Title:    note.GetTitle(),
				Desc:     note.GetDesc(),
				CreateAt: note.GetCreateAt(),
				UpdateAt: note.GetUpdateAt(),
				Author: &searchv1.Note_Author{
					Uid:      note.GetOwner(),
					Nickname: metadata.UserNickname(ctx),
				},
				TagList:       tagList,
				Visibility:    vis,
				AssetType:     assetType,
				LikesCount:    note.Likes,
				CommentsCount: note.Replies,
			}}

			_, err := dep.DocumentServer().BatchAddNote(ctx, &searchv1.BatchAddNoteRequest{Notes: docNote})
			if err != nil {
				xlog.Msg("creator sync note to searcher failed").Err(err).Extras("note_id", noteId).Errorx(ctx)
				return xerror.Wrapf(err, "batch add note failed").WithExtra("note_id", noteId).WithCtx(ctx)
			}

			return nil
		},
	})
}

// AsyncDeleteNoteFromSearcher 从搜索引擎删除笔记
func (b *Biz) AsyncDeleteNoteFromSearcher(ctx context.Context, noteId int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: fmt.Sprintf("creator_unsync_note_%d", noteId),
		Job: func(ctx context.Context) error {
			_, err := dep.DocumentServer().BatchDeleteNote(ctx, &searchv1.BatchDeleteNoteRequest{
				Ids: []string{imodel.NoteId(noteId).String()},
			})
			if err != nil {
				xlog.Msg("creator unsync note to searcher failed").Err(err).Extras("note_id", noteId).Errorx(ctx)
				return err
			}
			return nil
		},
	})
}
