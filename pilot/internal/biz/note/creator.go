package note

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/storage"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/note/model"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
	modelerr "github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"
)

// GetNote 获取笔记
func (b *Biz) GetNote(ctx context.Context, noteId int64) (*notev1.NoteItem, error) {
	curNote, err := dep.NoteCreatorServer().GetNote(ctx,
		&notev1.GetNoteRequest{NoteId: noteId})
	if err != nil {
		return nil, xerror.Wrapf(err, "get note failed").WithExtra("note_id", noteId).WithCtx(ctx)
	}

	return curNote.GetNote(), nil
}

// CreateNote 创建笔记
func (b *Biz) CreateNote(ctx context.Context, req *notev1.CreateNoteRequest) (*model.CreateNoteRes, error) {
	resp, err := dep.NoteCreatorServer().CreateNote(ctx, req)
	if err != nil {
		return nil, err
	}

	return &model.CreateNoteRes{NoteId: imodel.NoteId(resp.NoteId)}, nil
}

// UpdateNote 更新笔记
func (b *Biz) UpdateNote(ctx context.Context, noteId imodel.NoteId, req *notev1.CreateNoteRequest) error {
	_, err := dep.NoteCreatorServer().UpdateNote(ctx, &notev1.UpdateNoteRequest{
		NoteId: int64(noteId),
		Note:   req,
	})
	return err
}

// DeleteNote 删除笔记
func (b *Biz) DeleteNote(ctx context.Context, noteId imodel.NoteId) error {
	_, err := dep.NoteCreatorServer().DeleteNote(ctx, &notev1.DeleteNoteRequest{NoteId: int64(noteId)})
	return err
}

// PageListNotes 分页列出笔记
func (b *Biz) PageListNotes(ctx context.Context, page, count int32) (*notev1.PageListNoteResponse, error) {
	return dep.NoteCreatorServer().PageListNote(ctx, &notev1.PageListNoteRequest{
		Page:  page,
		Count: count,
	})
}

// AfterNoteUpserted 笔记创建/更新后的处理
func (b *Biz) AfterNoteUpserted(ctx context.Context, note *notev1.NoteItem) {
	b.AsyncNoteToSearcher(ctx, note.NoteId, note)
	b.NotifyWhenAtUsers(ctx, note)
	b.AppendRecentContacts(ctx, note)
}

// MarkNoteImages 检查笔记图片
func (b *Biz) MarkNoteImages(ctx context.Context, images []model.NoteImage) error {
	var (
		keys        = make([]string, 0, len(images))
		bucket, key string
		err         error
	)

	for _, img := range images {
		bucket, key, err = b.storageBiz.SeperateResource(uploadresource.NoteImage, img.FileId)
		if err != nil {
			return err
		}
		keys = append(keys, key)
	}

	_, err = b.storageBiz.BatchCheckResourceExist(ctx, bucket, keys, true)
	if err != nil {
		if errors.Is(err, bizstorage.ErrResourceNotFound) {
			err = modelerr.ErrResourceNotFound
		}
	} else {
		shouldTag := true
		for _, key := range keys {
			var (
				content []byte
				total   int64
			)
			content, total, err = b.storageBiz.GetResourceBytes(ctx, bucket, key, 32)
			if err != nil {
				if errors.Is(err, bizstorage.ErrResourceNotFound) {
					err = modelerr.ErrResourceNotFound
				}
				shouldTag = false
				break
			}

			if err = uploadresource.NoteImage.Check(content, total); err != nil {
				shouldTag = false
				break
			}
		}

		if shouldTag {
			b.storageBiz.BatchMarkResourceActive(ctx, bucket, keys, false)
		}
	}

	return err
}

func (b *Biz) UnmarkNoteImages(ctx context.Context, images []model.NoteImage) error {
	var (
		keys        = make([]string, 0, len(images))
		bucket, key string
		err         error
	)

	for _, img := range images {
		bucket, key, err = b.storageBiz.SeperateResource(uploadresource.NoteImage, img.FileId)
		if err != nil {
			return err
		}
		keys = append(keys, key)
	}
	b.storageBiz.BatchMarkResourceInactive(ctx, bucket, keys, false)

	return nil
}

// CheckNoteVideo 检查笔记视频
func (b *Biz) CheckNoteVideo(ctx context.Context) error {
	// TODO
	return nil
}

// NotifyWhenAtUsers 笔记中@用户通知
func (b *Biz) NotifyWhenAtUsers(ctx context.Context, note *notev1.NoteItem) {
	if IsNotePrivate(note) {
		return
	}

	if len(note.AtUsers) == 0 {
		return
	}
	uid := metadata.Uid(ctx)

	atUids := make([]int64, 0, len(note.AtUsers))
	for _, atUser := range note.AtUsers {
		atUids = append(atUids, atUser.Uid)
	}
	atUids = xslice.Uniq(atUids)

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "note_creator_notify_at_users",
		Job: func(ctx context.Context) error {
			atUsers, err := b.userBiz.ListUsersV2(ctx, atUids)
			if err != nil {
				xlog.Msg("note creator user biz list users failed").Err(err).Errorx(ctx)
				return err
			}

			if len(atUsers) == 0 {
				xlog.Msg("user biz return empty at users").Errorx(ctx)
				return nil
			}

			validNoteAtUsers := make([]*notev1.NoteAtUser, 0, len(note.AtUsers))
			for _, noteAtUser := range note.AtUsers {
				if _, ok := atUsers[noteAtUser.Uid]; ok {
					validNoteAtUsers = append(validNoteAtUsers, noteAtUser)
				}
			}

			err = b.notifyBiz.NotifyAtUsersOnNote(ctx, &bizsysnotify.NotifyAtUsersOnNoteReq{
				Uid:         uid,
				TargetUsers: validNoteAtUsers,
				Content: &bizsysnotify.NotifyAtUsersOnNoteReqContent{
					NoteDesc:  note.Desc,
					NoteId:    imodel.NoteId(note.NoteId),
					SourceUid: uid,
				},
			})
			if err != nil {
				xlog.Msg("note creator notify biz failed to notify at users").Err(err).Errorx(ctx)
				return err
			}

			return nil
		},
	})
}

// AppendRecentContacts 追加最近联系人
func (b *Biz) AppendRecentContacts(ctx context.Context, note *notev1.NoteItem) {
	uid := metadata.Uid(ctx)

	atUsers := imodel.AtUserList{}
	for _, atUser := range note.GetAtUsers() {
		if atUser.Uid == metadata.Uid(ctx) {
			continue
		}
		atUsers = append(atUsers, imodel.AtUser{
			Uid:      atUser.Uid,
			Nickname: atUser.Nickname,
		})
	}

	b.userBiz.AsyncAppendRecentContactsAtUser(ctx, uid, atUsers)
}
