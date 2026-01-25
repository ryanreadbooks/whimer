package notecreator

import (
	"context"

	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator/dto"
	commentrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/repository"
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage"
	storagevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	noterepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	"golang.org/x/sync/errgroup"
)

// 只负责业务逻辑编排
type Service struct {
	noteCreatorAdapter noterepo.NoteCreatorAdapter
	noteLikesAdapter   noterepo.NoteLikesAdapter
	commentAdapter     commentrepo.CommentAdapter
	storageRepository  storage.Repository
}

func NewService(
	noteCreatorAdapter noterepo.NoteCreatorAdapter,
	noteLikesAdapter noterepo.NoteLikesAdapter,
	commentAdapter commentrepo.CommentAdapter,
	storageRepository storage.Repository,
) *Service {
	return &Service{
		noteCreatorAdapter: noteCreatorAdapter,
		noteLikesAdapter:   noteLikesAdapter,
		commentAdapter:     commentAdapter,
		storageRepository:  storageRepository,
	}
}

func convertCommandToCreateNoteParams(cmd *dto.CreateNoteCommand) *entity.CreateNoteParams {
	params := &entity.CreateNoteParams{
		Title:     cmd.Basic.Title,
		Desc:      cmd.Basic.Desc,
		Privacy:   cmd.Basic.Privacy,
		AssetType: cmd.Basic.Type.AsAssetType(),
	}

	// images
	for _, img := range cmd.Images {
		params.Images = append(params.Images, &entity.NoteImage{
			FileId: img.FileId,
			Width:  img.Width,
			Height: img.Height,
			Format: img.Format,
		})
	}

	// video
	if cmd.Video != nil {
		params.Videos = &entity.NoteVideo{
			FileId:      cmd.Video.FileId,
			CoverFileId: cmd.Video.CoverFileId,
		}
	}

	// tagList
	for _, tag := range cmd.TagList {
		params.Tags = append(params.Tags, &entity.NoteTag{Id: int64(tag.Id)})
	}

	// atUsers
	for _, user := range cmd.AtUsers {
		params.AtUsers = append(params.AtUsers, mentionvo.AtUser{
			Uid:      user.Uid,
			Nickname: user.Nickname,
		})
	}

	return params
}

// 创作者发布一篇笔记流程
func (s *Service) CreateNote(ctx context.Context, cmd *dto.CreateNoteCommand) (
	*dto.CreateNoteResult, error,
) {
	err := s.preHandleNoteAsset(ctx, cmd)
	if err != nil {
		return nil, err
	}

	params := convertCommandToCreateNoteParams(cmd)
	err = s.preHandleNoteVideo(cmd.Basic.Type, cmd.Video, params)
	if err != nil {
		return nil, err
	}

	noteId, err := s.noteCreatorAdapter.CreateNote(ctx, params)
	if err != nil {
		return nil, err
	}

	return &dto.CreateNoteResult{NoteId: notevo.NoteId(noteId)}, nil
}

func (s *Service) UpdateNote(ctx context.Context, cmd *dto.UpdateNoteCommand) error {
	err := s.preHandleNoteAsset(ctx, &cmd.CreateNoteCommand)
	if err != nil {
		return err
	}

	params := convertCommandToCreateNoteParams(&cmd.CreateNoteCommand)
	err = s.preHandleNoteVideo(cmd.Basic.Type, cmd.Video, params)
	if err != nil {
		return err
	}
	_, err = s.noteCreatorAdapter.UpdateNote(ctx, &entity.UpdateNoteParams{
		NoteId:           int64(cmd.NoteId),
		CreateNoteParams: *params,
	})
	if err != nil {
		return err
	}

	return nil
}

// 删除笔记
func (s *Service) DeleteNote(ctx context.Context, noteId notevo.NoteId) error {
	// get first
	id := int64(noteId)
	_, err := s.noteCreatorAdapter.GetNote(ctx, id)
	if err != nil {
		return err
	}

	err = s.noteCreatorAdapter.DeleteNote(ctx, id)
	if err != nil {
		return err
	}

	// TODO unmark note assets

	return nil
}

func (s *Service) preHandleNoteAsset(ctx context.Context, cmd *dto.CreateNoteCommand) error {
	switch cmd.Basic.Type {
	case notevo.NoteTypeImage:
		return s.checkAndMarkNoteImages(ctx, cmd.Images)
	case notevo.NoteTypeVideo:
		return s.checkAndMarkNoteVideo(ctx, cmd.Video)
	}
	return nil
}

func (s *Service) preHandleNoteVideo(noteType notevo.NoteType,
	video *dto.VideoForCreateNote, params *entity.CreateNoteParams,
) error {
	if noteType.IsVideo() {
		// 比如： fildId = videos/note/cosmic/123.mp4
		//  rawKey = 123.mp4
		//  meta.PrefixSegment = note
		//  targetId = videos/note/123.mp4
		rawKey := s.storageRepository.TrimBucketAndPrefix(storagevo.ObjectTypeNoteVideo, video.FileId)
		meta, err := s.storageRepository.GetObjectMeta(storagevo.ObjectTypeNoteVideo)
		if err != nil {
			return err
		}
		targetId := meta.Bucket + "/" + meta.PrefixSegment + "/" + rawKey
		params.SetVideoTargetFileId(targetId)
	}
	return nil
}

func (s *Service) checkAndMarkNoteImages(ctx context.Context, images dto.ImageListForCreateNote) error {
	imgIds := make([]storagevo.ObjectInfo, 0, len(images))
	for _, img := range images {
		imgIds = append(imgIds, storagevo.ObjectInfo{
			FileId: img.FileId,
		})
	}
	err := s.storageRepository.CheckAndMarkObjects(ctx, storagevo.ObjectTypeNoteImage, imgIds)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) checkAndMarkNoteVideo(ctx context.Context, video *dto.VideoForCreateNote) error {
	if video.FileId != "" {
		err := s.storageRepository.CheckAndMarkObjects(ctx,
			storagevo.ObjectTypeNoteVideo,
			[]storagevo.ObjectInfo{{FileId: video.FileId}})
		if err != nil {
			return err
		}
	}

	if video.CoverFileId != "" {
		err := s.storageRepository.CheckAndMarkObjects(ctx,
			storagevo.ObjectTypeNoteVideoCover,
			[]storagevo.ObjectInfo{{FileId: video.CoverFileId}})
		if err != nil {
			return err
		}
	}

	return nil
}

// 分页查询笔记
func (s *Service) PageListNotes(ctx context.Context, query *dto.PageListNotesQuery) (
	*dto.NoteList, error,
) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	result, err := s.noteCreatorAdapter.PageListNotes(ctx,
		&entity.PageListNotesParams{
			Page:   query.Page,
			Count:  query.Count,
			Status: query.Status,
		})
	if err != nil {
		return nil, err
	}

	uid := metadata.Uid(ctx)

	noteItems := make([]*dto.Note, 0, len(result.Items))
	noteIds := make([]int64, 0, len(result.Items))
	for _, item := range result.Items {
		dtoNote := s.convertEntityNoteToDto(ctx, item)
		ipLoc, _ := infra.Ip2Loc().Convert(ctx, item.Ip)
		dtoNote.IpLoc = ipLoc
		noteItems = append(noteItems, dtoNote)
		noteIds = append(noteIds, item.Id.Int64())
	}

	// 填充interacts
	interacts, _ := s.batchGetNoteInteraction(ctx, uid, noteIds)
	for _, noteItem := range noteItems {
		noteItem.Interact = interacts[noteItem.NoteId.Int64()]
	}

	return &dto.NoteList{
		Total: result.Total,
		Items: noteItems,
	}, nil
}

// 获取单个笔记
func (s *Service) GetNote(ctx context.Context, noteId notevo.NoteId) (*dto.Note, error) {
	note, err := s.noteCreatorAdapter.GetNote(ctx, int64(noteId))
	if err != nil {
		return nil, err
	}

	uid := metadata.Uid(ctx)
	dtoNote := s.convertEntityNoteToDto(ctx, note)
	interact, _ := s.getNoteInteraction(ctx, uid, note.Id.Int64())
	dtoNote.Interact = interact
	ipLoc, _ := infra.Ip2Loc().Convert(ctx, note.Ip)
	dtoNote.IpLoc = ipLoc

	return dtoNote, nil
}

func (s *Service) convertEntityNoteToDto(ctx context.Context, note *entity.CreatorNote) *dto.Note {
	if note == nil {
		return nil
	}

	atUsers := make([]*commondto.AtUser, 0, len(note.AtUsers))
	tagList := make([]*commondto.NoteTag, 0, len(note.Tags))
	images := make(commondto.NoteImageList, 0, len(note.Images))
	videos := make(commondto.NoteVideoList, 0, len(note.Videos))

	for _, img := range note.Images {
		images = append(images, &commondto.NoteImage{
			Key:        img.FileId,
			Url:        commondto.NewNoteImageUrl(img.FileId),
			UrlPreview: commondto.NewNoteImagePreviewUrl(img.FileId),
			Type:       notevo.AssetTypeImage,
			Metadata: commondto.NoteImageMetadata{
				Format: img.Format,
				Width:  img.Width,
				Height: img.Height,
			},
		})
	}

	for _, video := range note.Videos {
		if video != nil {
			url, _ := s.storageRepository.PresignGetUrl(ctx, storagevo.ObjectTypeNoteVideo, video.FileId)
			videos = append(videos, &commondto.NoteVideo{
				Key:  video.FileId,
				Type: notevo.AssetTypeVideo,
				Url:  url,
				Metadata: commondto.NoteVideoMeta{
					Width:      video.GetMetadata().Width,
					Height:     video.GetMetadata().Height,
					Format:     video.GetMetadata().Format,
					Duration:   video.GetMetadata().Duration,
					Bitrate:    video.GetMetadata().Bitrate,
					Codec:      video.GetMetadata().Codec,
					Framerate:  video.GetMetadata().Framerate,
					AudioCodec: video.GetMetadata().AudioCodec,
				},
			})
		}
	}

	dn := &dto.Note{
		NoteId:   note.Id,
		Title:    note.Title,
		Desc:     note.Desc,
		Privacy:  int8(note.Privacy),
		CreateAt: note.CreateTime,
		UpdateAt: note.UpdateTime,
		Type:     note.Type,

		Likes:   note.Likes,
		Replies: note.Replies,
		AtUsers: atUsers,
		TagList: tagList,

		Images: images,
		Videos: videos,
	}

	return dn
}

func (s *Service) getNoteInteraction(ctx context.Context, uid, noteId int64) (
	commondto.NoteInteraction, error,
) {
	var interact commondto.NoteInteraction
	eg := errgroup.Group{}
	eg.Go(func() error {
		return recovery.Do(func() error {
			likeStatus, err := s.noteLikesAdapter.GetLikeStatus(ctx,
				&noterepo.GetLikeStatusParams{
					Uid:    uid,
					NoteId: noteId,
				})
			if err != nil {
				return err
			}

			interact.Liked = likeStatus.Liked

			return nil
		})
	})

	eg.Go(func() error {
		return recovery.Do(func() error {
			commentStatus, err := s.commentAdapter.CheckCommented(ctx,
				&commentrepo.CheckCommentedParams{
					Uid:     uid,
					NoteIds: []int64{noteId},
				})
			if err != nil {
				return err
			}

			interact.Commented = commentStatus.Commented[uid]
			return nil
		})
	})

	err := eg.Wait()
	if err != nil {
		// log only
		xlog.Msgf("get note interact failed").Extras("note_id", noteId).Errorx(ctx)
	}

	return interact, nil
}

func (s *Service) batchGetNoteInteraction(ctx context.Context, uid int64, noteIds []int64) (
	map[int64]commondto.NoteInteraction, error,
) {
	var (
		eg            = errgroup.Group{}
		likeStatus    *noterepo.BatchGetLikeStatusResult
		commentStatus *commentrepo.CheckCommentedResult
	)

	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			likeStatus, err = s.noteLikesAdapter.BatchGetLikeStatus(ctx,
				&noterepo.BatchGetLikeStatusParams{
					Uid:     uid,
					NoteIds: noteIds,
				},
			)
			if err != nil {
				return err
			}

			return nil
		})
	})

	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			commentStatus, err = s.commentAdapter.CheckCommented(ctx,
				&commentrepo.CheckCommentedParams{
					Uid:     uid,
					NoteIds: noteIds,
				})
			if err != nil {
				return err
			}

			return nil
		})
	})

	err := eg.Wait()
	if err != nil {
		xlog.Msgf("batch get not interact failed").Extras("note_ids", noteIds).Errorx(ctx)
	}

	m := make(map[int64]commondto.NoteInteraction)
	for _, noteId := range noteIds {
		liked := false
		commented := false
		if likeStatus != nil {
			liked = likeStatus.Liked[noteId]
		}
		if commentStatus != nil {
			commented = commentStatus.Commented[noteId]
		}

		m[noteId] = commondto.NoteInteraction{
			Liked:     liked,
			Commented: commented,
		}
	}

	return m, nil
}

// AddTag 新增标签
func (s *Service) AddTag(ctx context.Context, name string) (*dto.AddTagResult, error) {
	tagId, err := s.noteCreatorAdapter.AddTag(ctx, name)
	if err != nil {
		return nil, err
	}
	return &dto.AddTagResult{TagId: notevo.TagId(tagId)}, nil
}

// SearchTags 搜索标签
func (s *Service) SearchTags(ctx context.Context, name string) ([]*dto.SearchedTag, error) {
	tags, err := s.noteCreatorAdapter.SearchTags(ctx, name)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.SearchedTag, 0, len(tags))
	for _, tag := range tags {
		result = append(result, &dto.SearchedTag{
			Id:   tag.Id,
			Name: tag.Name,
		})
	}
	return result, nil
}
