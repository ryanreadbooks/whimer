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
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
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

func (b *Biz) treatNoteVideoReq(req *notev1.CreateNoteRequest) {
	// 视频需要特殊处理 req
	// 比如： fildId = videos/note/cosmic/123.mp4
	//  rawKey = 123.mp4
	//  meta.PrefixSegment = note
	//  targetId = video/note/123.mp4
	if req.Basic.AssetType == notev1.NoteAssetType_VIDEO {
		rawKey := b.storageBiz.TrimBucketAndPrefix(uploadresource.NoteVideo, req.Video.FileId)
		meta := config.Conf.UploadResourceDefineMap[uploadresource.NoteVideo]
		targetId := meta.Bucket + "/" + meta.PrefixSegment + "/" + rawKey
		req.Video.TargetFileId = targetId
	}
}

// CreateNote 创建笔记
func (b *Biz) CreateNote(ctx context.Context, req *notev1.CreateNoteRequest) (*model.CreateNoteRes, error) {
	b.treatNoteVideoReq(req)
	resp, err := dep.NoteCreatorServer().CreateNote(ctx, req)
	if err != nil {
		return nil, err
	}

	return &model.CreateNoteRes{NoteId: imodel.NoteId(resp.NoteId)}, nil
}

// UpdateNote 更新笔记
func (b *Biz) UpdateNote(ctx context.Context, noteId imodel.NoteId, req *notev1.CreateNoteRequest) error {
	if req.Video != nil && req.Video.FileId != "" {
		b.treatNoteVideoReq(req)
	}
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
func (b *Biz) PageListNotes(ctx context.Context, page, count int32, status model.NoteStatus) (*notev1.PageListNoteResponse, error) {
	return dep.NoteCreatorServer().PageListNote(ctx, &notev1.PageListNoteRequest{
		Page:           page,
		Count:          count,
		LifeCycleState: model.NoteStatusAsPb(status),
	})
}

// AfterNoteUpserted 笔记创建/更新后的处理
func (b *Biz) AfterNoteUpserted(ctx context.Context, note *notev1.NoteItem) {
	// b.AsyncNoteToSearcher(ctx, note.NoteId, note) // 改为note_event处理
	b.NotifyWhenAtUsers(ctx, note)
	b.AppendRecentContacts(ctx, note)
}

// resourceChecker 资源检查器（可选）
type resourceChecker func(ctx context.Context, bucket string, keys []string) error

// collectResourceKeys 收集资源的 bucket 和 keys
func (b *Biz) collectResourceKeys(
	resType uploadresource.Type,
	fileIds []string,
) (bucket string, keys []string, err error) {
	keys = make([]string, 0, len(fileIds))
	for _, fileId := range fileIds {
		var key string
		bucket, key, err = b.storageBiz.SeperateResource(resType, fileId)
		if err != nil {
			return "", nil, err
		}
		keys = append(keys, key)
	}
	return bucket, keys, nil
}

// checkAndMarkResources 检查并标记资源为激活状态
func (b *Biz) checkAndMarkResources(
	ctx context.Context,
	resType uploadresource.Type,
	fileIds []string,
	checker resourceChecker,
) error {
	if len(fileIds) == 0 {
		return nil
	}

	for _, fileId := range fileIds {
		err := b.storageBiz.CheckFileIdValid(resType, fileId)
		if err != nil {
			return err
		}
	}

	bucket, keys, err := b.collectResourceKeys(resType, fileIds)
	if err != nil {
		return err
	}

	_, err = b.storageBiz.BatchCheckResourceExist(ctx, bucket, keys, true)
	if err != nil {
		if errors.Is(err, bizstorage.ErrResourceNotFound) {
			return modelerr.ErrResourceNotFound
		}
		return err
	}

	if checker != nil {
		if err = checker(ctx, bucket, keys); err != nil {
			return err
		}
	}

	b.storageBiz.BatchMarkResourceActive(ctx, bucket, keys, false)
	return nil
}

// unmarkResources 取消标记资源
func (b *Biz) unmarkResources(ctx context.Context, resType uploadresource.Type, fileIds []string) error {
	if len(fileIds) == 0 {
		return nil
	}

	bucket, keys, err := b.collectResourceKeys(resType, fileIds)
	if err != nil {
		return err
	}

	b.storageBiz.BatchMarkResourceInactive(ctx, bucket, keys, false)
	return nil
}

// CheckAndMarkNoteImages 检查笔记图片
func (b *Biz) CheckAndMarkNoteImages(ctx context.Context, images []model.NoteImage) error {
	fileIds := make([]string, 0, len(images))
	for _, img := range images {
		fileIds = append(fileIds, img.FileId)
	}

	return b.checkAndMarkResources(ctx, uploadresource.NoteImage, fileIds, b.checkImageContent)
}

// checkImageContent 检查图片内容格式和大小
func (b *Biz) checkImageContent(ctx context.Context, bucket string, keys []string) error {
	for _, key := range keys {
		content, size, err := b.storageBiz.GetResourceBytes(ctx, bucket, key, 32)
		if err != nil {
			if errors.Is(err, bizstorage.ErrResourceNotFound) {
				return modelerr.ErrResourceNotFound
			}
			return err
		}
		if err = uploadresource.NoteImage.Check(content, size); err != nil {
			return err
		}
	}
	return nil
}

// UnmarkNoteImages 取消标记笔记图片
func (b *Biz) UnmarkNoteImages(ctx context.Context, images []model.NoteImage) error {
	fileIds := make([]string, 0, len(images))
	for _, img := range images {
		fileIds = append(fileIds, img.FileId)
	}
	return b.unmarkResources(ctx, uploadresource.NoteImage, fileIds)
}

func (b *Biz) checkVideoContent(ctx context.Context, bucket string, keys []string) error {
	for _, key := range keys {
		content, size, err := b.storageBiz.GetResourceBytes(ctx, bucket, key, 32)
		if err != nil {
			if errors.Is(err, bizstorage.ErrResourceNotFound) {
				return modelerr.ErrResourceNotFound
			}
			return err
		}
		if err = uploadresource.NoteVideo.Check(content, size); err != nil {
			return err
		}
	}

	return nil
}

// CheckAndMarkNoteVideo 检查笔记视频
func (b *Biz) CheckAndMarkNoteVideo(ctx context.Context, video model.NoteVideo) error {
	if video.FileId == "" {
		return modelerr.ErrResourceNotFound
	}
	return b.checkAndMarkResources(ctx, uploadresource.NoteVideo, []string{video.FileId}, b.checkVideoContent)
}

// UnmarkNoteVideo 取消标记笔记视频
func (b *Biz) UnmarkNoteVideo(ctx context.Context, video model.NoteVideo) error {
	if video.FileId == "" {
		return nil
	}
	return b.unmarkResources(ctx, uploadresource.NoteVideo, []string{video.FileId})
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
