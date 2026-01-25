package notefeed

import (
	"context"

	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notefeed/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notefeed/errors"
	commentrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage"
	storagevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	noterepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	relationrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/repository"
	userrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"

	"golang.org/x/sync/errgroup"
)

func isReqUserGuest(ctx context.Context) bool {
	return metadata.Uid(ctx) == 0
}

type Service struct {
	noteFeedAdapter     noterepo.NoteFeedAdapter
	noteInteractAdapter noterepo.NoteLikesAdapter
	noteSearchAdapter   noterepo.NoteSearchAdapter
	userAdapter         userrepo.UserServiceAdapter
	relationAdapter     relationrepo.RelationAdapter
	storageRepository   storage.Repository
	commentAdapter      commentrepo.CommentAdapter
	userSettingAdapter  userrepo.UserSettingRepository
}

func NewService(
	noteFeedAdapter noterepo.NoteFeedAdapter,
	noteInteractAdapter noterepo.NoteLikesAdapter,
	noteSearchAdapter noterepo.NoteSearchAdapter,
	userAdapter userrepo.UserServiceAdapter,
	relationAdapter relationrepo.RelationAdapter,
	storageRepository storage.Repository,
	commentAdapter commentrepo.CommentAdapter,
	userSettingAdapter userrepo.UserSettingRepository,
) *Service {
	return &Service{
		noteFeedAdapter:     noteFeedAdapter,
		noteInteractAdapter: noteInteractAdapter,
		noteSearchAdapter:   noteSearchAdapter,
		userAdapter:         userAdapter,
		relationAdapter:     relationAdapter,
		storageRepository:   storageRepository,
		commentAdapter:      commentAdapter,
		userSettingAdapter:  userSettingAdapter,
	}
}

func (s *Service) GetRandom(
	ctx context.Context,
	query *dto.GetRandomQuery,
) ([]*dto.FeedNote, error) {
	// 1. 获取笔记基础信息
	resp, err := s.noteFeedAdapter.RandomGet(ctx, int32(query.NeedNum))
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter random get failed").WithExtras("query", query).WithCtx(ctx)
	}

	feedNotes, err := s.assembleFeedNotes(ctx, resp)
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter assemble feed notes failed").WithCtx(ctx)
	}

	return feedNotes, nil
}

func (s *Service) GetFeedNote(
	ctx context.Context,
	query *dto.GetFeedNoteQuery,
) (*dto.FullFeedNote, error) {
	note, ext, err := s.noteFeedAdapter.GetNote(ctx, query.NoteId.Int64())
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter get note failed").WithExtras("query", query).WithCtx(ctx)
	}

	feedNotes, err := s.assembleFeedNotes(ctx, []*entity.FeedNote{note})
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter assemble feed notes failed").WithCtx(ctx)
	}

	targetNote := feedNotes[0]

	// ext转换
	tagList := make([]*commondto.NoteTag, 0, len(ext.Tags))
	for _, tag := range ext.Tags {
		tagList = append(tagList, &commondto.NoteTag{
			Id:   notevo.TagId(tag.Id),
			Name: tag.Name,
		})
	}

	// atUser
	atUsers := make([]*commondto.AtUser, 0, len(ext.AtUsers))
	for _, atUser := range ext.AtUsers {
		atUsers = append(atUsers, &commondto.AtUser{
			Uid:      atUser.Uid,
			Nickname: atUser.Nickname,
		})
	}

	return &dto.FullFeedNote{
		FeedNote: targetNote,
		TagList:  tagList,
		AtUsers:  atUsers,
	}, nil
}

// 获取用户的笔记
func (s *Service) ListUserFeedNotes(
	ctx context.Context,
	uid int64,
	cursor int64, count int32,
) ([]*dto.FeedNote, *commondto.PageResult, error) {
	notes, page, err := s.noteFeedAdapter.ListUserNote(ctx, uid, cursor, count)
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "note feed adapter list user note failed").
			WithExtras("cursor", cursor, "count", count).WithCtx(ctx)
	}

	feedNotes, err := s.assembleFeedNotes(ctx, notes)
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "note feed adapter assemble feed notes failed").WithCtx(ctx)
	}

	return feedNotes, &commondto.PageResult{
		NextCursor: page.NextCursor,
		HasNext:    page.HasNext,
	}, nil
}

// 列出用户点赞过的笔记
func (s *Service) ListUserLikedNotes(
	ctx context.Context,
	uid int64,
	cursor string, count int32,
) ([]*dto.FeedNote, *commondto.PageResultV2, error) {
	// 检查用户点赞的笔记是否公开
	uidSetting, err := s.userSettingAdapter.GetLocalSetting(ctx, uid)
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "user setting adapter get local setting failed").
			WithExtras("uid", uid).WithCtx(ctx)
	}

	operator := metadata.Uid(ctx)
	if operator != uid {
		if !uidSetting.ShouldShowNoteLikes() {
			return nil, nil, errors.ErrLikesHistoryHidden
		}
	}

	notes, page, err := s.noteFeedAdapter.ListUserLikedNote(ctx, uid, cursor, count)
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "note feed adapter list user liked note failed").
			WithExtras("cursor", cursor, "count", count).WithCtx(ctx)
	}

	feedNotes, err := s.assembleFeedNotes(ctx, notes)
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "note feed adapter assemble feed notes failed").WithCtx(ctx)
	}

	return feedNotes, &commondto.PageResultV2{
		NextCursor: page.NextCursor,
		HasNext:    page.HasNext,
	}, nil
}

// 组装笔记信息
func (s *Service) assembleFeedNotes(
	ctx context.Context,
	notes []*entity.FeedNote,
) ([]*dto.FeedNote, error) {
	var (
		err     error
		reqUid  = metadata.Uid(ctx)
		authors = make(map[int64][]int64, len(notes)) // 作者，一个作者可能对应多篇笔记
		noteIds = make([]int64, 0, len(notes))        // 全部笔记id
	)

	// 收集author uid
	for _, note := range notes {
		authors[note.AuthorUid] = append(authors[note.AuthorUid], note.Id.Int64())
		noteIds = append(noteIds, note.Id.Int64())
	}

	var (
		authorInfos  map[int64]*commondto.User // uid -> author info
		oidCommented map[int64]bool            // reqUid是否评论过noteid
		oidLiked     map[int64]bool            // reqUid是否点赞过noteid
		followStatus map[int64]bool            // reqUid是否关注了authorUid
	)

	eg := errgroup.Group{}
	authorUids := xslice.Uniq(xmap.Keys(authors))
	// 2. 获取各篇笔记的作者信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			authorInfos, err = s.collectAuthorInfos(ctx, authorUids)
			return err
		})
	})

	// 3. 获取各篇笔记当前用户的评论交互信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			oidCommented, err = s.collectCommentStatus(ctx, reqUid, noteIds)
			return err
		})
	})

	// 4. 获取各篇笔记当前用户的点赞交互信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			oidLiked, err = s.collectLikeStatus(ctx, reqUid, noteIds)
			return err
		})
	})

	// 5. 获取reqUid对笔记作者的关注状态
	authorUidsClean := make([]int64, len(authorUids))
	copy(authorUidsClean, authorUids)
	// authorUids中排除当前请求的reqUid
	authorUidsClean = xslice.Filter(authorUidsClean, func(_ int, v int64) bool { return v == reqUid })
	if len(authorUidsClean) != 0 {
		eg.Go(func() error {
			return recovery.Do(func() error {
				var err error
				followStatus, err = s.collectFollowStatus(ctx, reqUid, authorUids)
				return err
			})
		})
	}

	err = eg.Wait()
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter assemble feed notes failed").WithCtx(ctx)
	}

	// 组织最后结果
	feedNotes := make([]*dto.FeedNote, 0, len(notes))
	for _, note := range notes {
		feedNote := s.convertEntityToFeedNoteDto(ctx, note)
		// 填充author
		author := authorInfos[note.AuthorUid]
		if author != nil {
			feedNote.Author = author
		}

		// 填充interact
		feedNote.Interact.Commented = oidCommented[note.Id.Int64()]
		feedNote.Interact.Liked = oidLiked[note.Id.Int64()]
		feedNote.Interact.Following = followStatus[note.AuthorUid]

		feedNotes = append(feedNotes, feedNote)
	}

	return feedNotes, nil
}

func (s *Service) convertEntityToFeedNoteDto(ctx context.Context, note *entity.FeedNote) *dto.FeedNote {
	if note == nil {
		return nil
	}

	images := make(commondto.NoteImageList, 0, len(note.Images))
	for _, image := range note.Images {
		images = append(images, commondto.ConvertEntityNoteImageToDto(image))
	}
	videos := make(commondto.NoteVideoList, 0, len(note.Videos))
	for _, video := range note.Videos {
		videoUrl, _ := s.storageRepository.PresignGetUrl(ctx, storagevo.ObjectTypeNoteVideo, video.FileId)
		videos = append(videos, commondto.ConvertEntityNoteVideoToDto(video, videoUrl))
	}

	ipLoc, _ := infra.Ip2Loc().Convert(ctx, note.Ip)
	n := &dto.FeedNote{
		NoteId:   note.Id,
		Title:    note.Title,
		Desc:     note.Desc,
		Type:     note.Type,
		CreateAt: note.CreateAt,
		UpdateAt: note.UpdateAt,
		IpLoc:    ipLoc,
		Likes:    note.Likes,
		Comments: note.Comments,

		Images: images,
		Videos: videos,
	}

	return n
}

func (s *Service) collectAuthorInfos(ctx context.Context, uids []int64) (map[int64]*commondto.User, error) {
	authorInfos, err := s.userAdapter.BatchGetUser(ctx, uids)
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter batch get authors failed").
			WithExtras("uids", uids).WithCtx(ctx)
	}

	users := make(map[int64]*commondto.User, len(authorInfos))
	for uid, user := range authorInfos {
		users[uid] = &commondto.User{
			Uid:      uid,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		}
	}

	return users, nil
}

// TODO 评论交互信息
func (s *Service) collectCommentStatus(ctx context.Context, reqUid int64, noteIds []int64) (map[int64]bool, error) {
	if isReqUserGuest(ctx) {
		return make(map[int64]bool), nil
	}

	commentStatus, err := s.commentAdapter.BatchCheckCommented(ctx,
		&commentrepo.BatchCheckCommentedParams{
			Uid:     reqUid,
			NoteIds: noteIds,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter batch get comment status failed").
			WithExtras("reqUid", reqUid).
			WithExtras("noteIds", noteIds).WithCtx(ctx)
	}

	return commentStatus.Commented, nil
}

// 点赞交互信息
func (s *Service) collectLikeStatus(ctx context.Context, reqUid int64, noteIds []int64) (map[int64]bool, error) {
	if isReqUserGuest(ctx) {
		return make(map[int64]bool), nil
	}

	likeStatuses, err := s.noteInteractAdapter.BatchGetLikeStatus(ctx,
		&noterepo.BatchGetLikeStatusParams{
			Uid:     reqUid,
			NoteIds: noteIds,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter batch get like status failed").
			WithExtras("reqUid", reqUid).WithExtras("noteIds", noteIds).WithCtx(ctx)
	}

	return likeStatuses.Liked, nil
}

// 获取关注关系
func (s *Service) collectFollowStatus(ctx context.Context, reqUid int64, authorUids []int64) (map[int64]bool, error) {
	if isReqUserGuest(ctx) {
		return make(map[int64]bool), nil
	}

	followStatuses, err := s.relationAdapter.BatchGetFollowStatus(ctx, reqUid, authorUids)
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter batch get follow status failed").
			WithExtras("reqUid", reqUid).WithExtras("authorUids", authorUids).WithCtx(ctx)
	}

	return followStatuses, nil
}

// 搜索笔记
func (s *Service) SearchNotes(ctx context.Context, query *dto.SearchNotesQuery) (*dto.SearchNotesResult, error) {
	// 1. 转换搜索过滤器
	filters := make([]*noterepo.NoteSearchFilter, 0, len(query.Filters))
	for _, f := range query.Filters {
		filters = append(filters, &noterepo.NoteSearchFilter{
			Type:  f.Type,
			Value: f.Value,
		})
	}

	// 2. 调用搜索适配器
	searchResult, err := s.noteSearchAdapter.SearchNote(ctx, &noterepo.SearchNoteParams{
		Keyword:   query.Keyword,
		PageToken: query.PageToken,
		Count:     query.Count,
		Filters:   filters,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "note search adapter search note failed").
			WithExtras("keyword", query.Keyword).WithCtx(ctx)
	}

	result := &dto.SearchNotesResult{
		NextToken: searchResult.NextToken,
		HasNext:   searchResult.HasNext,
		Total:     searchResult.Total,
	}

	if len(searchResult.NoteIds) == 0 {
		return result, nil
	}

	// 3. 获取笔记详情
	noteIds := make([]int64, 0, len(searchResult.NoteIds))
	for _, id := range searchResult.NoteIds {
		noteIds = append(noteIds, id.Int64())
	}

	notesMap, err := s.noteFeedAdapter.BatchGetNotes(ctx, noteIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "note feed adapter batch get notes failed").
			WithExtras("noteIds", noteIds).WithCtx(ctx)
	}

	// 按搜索结果顺序组装笔记列表
	notes := make([]*entity.FeedNote, 0, len(searchResult.NoteIds))
	for _, id := range searchResult.NoteIds {
		if note, ok := notesMap[id.Int64()]; ok {
			notes = append(notes, note)
		}
	}

	// 4. 组装 feed notes
	feedNotes, err := s.assembleFeedNotes(ctx, notes)
	if err != nil {
		return nil, xerror.Wrapf(err, "assemble feed notes failed").WithCtx(ctx)
	}

	result.Items = feedNotes
	return result, nil
}
