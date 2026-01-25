package feed

import (
	"context"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/feed/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"

	"golang.org/x/sync/errgroup"
)

type Biz struct {
}

func NewBiz() *Biz { return &Biz{} }

func isGuestFromCtx(ctx context.Context) bool {
	return imodel.IsGuestFromCtx(ctx)
}

// 收集作者信息
func (b *Biz) collectAuthor(ctx context.Context, uids []int64) (map[int64]*userv1.UserInfo, error) {
	authors := make(map[int64]*userv1.UserInfo)
	resp, err := dep.Userer().BatchGetUser(ctx, &userv1.BatchGetUserRequest{
		Uids: uids,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz failed to get user infos").WithExtra("uids", uids).WithCtx(ctx)
	}

	for _, u := range resp.GetUsers() {
		authors[u.Uid] = u
	}

	return authors, nil
}

// 收集reqUid和noteIds之间的评论关系
func (b *Biz) collectCommentStatus(ctx context.Context, reqUid int64, noteIds []int64) (map[int64]bool, error) {
	if isGuestFromCtx(ctx) {
		return map[int64]bool{}, nil
	}

	oidCommented := make(map[int64]bool)
	// uid -> [oid1, oid2, ...]
	objs := make(map[int64]*commentv1.BatchCheckUserOnObjectRequest_Objects)
	objs[reqUid] = &commentv1.BatchCheckUserOnObjectRequest_Objects{
		Oids: noteIds,
	}
	resp, err := dep.Commenter().BatchCheckUserOnObject(ctx,
		&commentv1.BatchCheckUserOnObjectRequest{
			Mappings: objs,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz failed to check user on object").WithCtx(ctx)
	}

	// organize result
	pairs := resp.GetResults()
	for _, comInfo := range pairs[reqUid].GetList() {
		oidCommented[comInfo.Oid] = comInfo.Commented
	}

	return oidCommented, nil
}

// 获取评论数量
func (b *Biz) collectCommentNumber(ctx context.Context, noteIds []int64) (map[int64]int64, error) {
	resp, err := dep.Commenter().BatchCountComment(ctx, &commentv1.BatchCountCommentRequest{
		Oids: noteIds,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz failed to count comment").WithCtx(ctx)
	}

	commentNums := resp.GetNumbers()

	return commentNums, nil
}

// 获取reqUid和noteIds之间的点赞关系
func (b *Biz) collectLikeStatus(ctx context.Context, reqUid int64, noteIds []int64) (map[int64]bool, error) {
	if isGuestFromCtx(ctx) {
		return make(map[int64]bool), nil
	}

	oidLiked := make(map[int64]bool)
	mappings := make(map[int64]*notev1.NoteIdList)
	mappings[reqUid] = &notev1.NoteIdList{NoteIds: noteIds}
	resp, err := dep.NoteInteractServer().BatchCheckUserLikeStatus(ctx,
		&notev1.BatchCheckUserLikeStatusRequest{
			Mappings: mappings,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz failed to batch check user like status").WithCtx(ctx)
	}

	pairs := resp.GetResults()
	for _, likedInfo := range pairs[reqUid].GetList() {
		oidLiked[likedInfo.NoteId] = likedInfo.Liked
	}

	return oidLiked, nil
}

func (b *Biz) collectRelationStatus(ctx context.Context, reqUid int64, authorUids []int64) (map[int64]bool, error) {
	if isGuestFromCtx(ctx) {
		return make(map[int64]bool), nil
	}

	resp, err := dep.RelationServer().BatchCheckUserFollowed(ctx,
		&relationv1.BatchCheckUserFollowedRequest{
			Uid:     reqUid,
			Targets: authorUids,
		})
	if err != nil {
		return nil, xerror.Wrapf(err,
			"feed biz failed to batch check user following authors status").WithCtx(ctx)
	}

	return resp.Status, nil
}

// 组装notes的各种字段
func (b *Biz) AssembleNoteFeeds(ctx context.Context, notes []*notev1.FeedNoteItem) (
	[]*model.FeedNoteItem, error) {
	var (
		err     error
		reqUid  = metadata.Uid(ctx)
		authors = make(map[int64][]int64, len(notes)) // 作者，一个作者可能对应多篇笔记
	)

	for _, note := range notes {
		authors[note.Author] = append(authors[note.Author], note.NoteId)
	}

	var (
		authorInfos  map[int64]*userv1.UserInfo // uid -> author info
		oidCommented map[int64]bool             // oid -> reqUid commented or not
		oidLiked     map[int64]bool             // oid -> reqUid liked or not
		commentNums  map[int64]int64            // oid -> comment count
		userFollows  map[int64]bool             // authorId -> isFollowed
	)

	noteIds := make([]int64, 0, len(notes)) // 全部笔记id
	for _, n := range notes {
		noteIds = append(noteIds, n.NoteId)
	}

	eg, ctx := errgroup.WithContext(ctx)
	authorUids := xslice.Uniq(xmap.Keys(authors))
	// 2. 获取各篇笔记的作者信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			uids := authorUids
			var err error
			authorInfos, err = b.collectAuthor(ctx, uids)
			return err
		})
	})

	// 3. 获取各篇笔记当前用户的交互信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			oidCommented, err = b.collectCommentStatus(ctx, reqUid, noteIds)
			return err
		})
	})

	// 4. 获取评论数量
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			commentNums, err = b.collectCommentNumber(ctx, noteIds)
			return err
		})
	})

	// 5. 获取reqUid对笔记的点赞
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			oidLiked, err = b.collectLikeStatus(ctx, reqUid, noteIds)
			return err
		})
	})

	// 6. 获取reqUid对笔记作者的关注状态
	authorUidsClean := make([]int64, len(authorUids))
	copy(authorUidsClean, authorUids)
	// authorUids中排除当前请求的reqUid
	authorUidsClean = xslice.Filter(authorUidsClean, func(_ int, v int64) bool { return v == reqUid })

	if len(authorUidsClean) != 0 {
		eg.Go(func() error {
			return recovery.Do(func() error {
				var err error
				userFollows, err = b.collectRelationStatus(ctx, reqUid, authorUidsClean)
				if err != nil {
					xlog.Msg("feed biz failed to collect relation status").Extras("authors", authorUidsClean).Err(err).Errorx(ctx)
				}

				// 非关键数据降级处理
				if userFollows == nil {
					userFollows = make(map[int64]bool)
				}

				return nil
			})
		})
	}

	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	// organize final results
	feedNotes := make([]*model.FeedNoteItem, 0, len(notes))
	for _, note := range notes {
		feedNote := model.NewFeedNoteItemFromPb(note)
		author := authorInfos[note.Author]
		if author != nil {
			feedNote.Author = model.NewAuthor(author)
		}

		noteId := int64(feedNote.NoteId)
		// fill extra fields
		feedNote.Comments = commentNums[noteId]
		feedNote.Interact.Commented = oidCommented[noteId]
		feedNote.Interact.Liked = oidLiked[noteId]
		feedNote.Interact.Following = userFollows[author.Uid]

		feedNotes = append(feedNotes, feedNote)
	}

	return feedNotes, nil
}

func (b *Biz) BatchGetNote(ctx context.Context, noteIds []int64) ([]*model.FeedNoteItem, error) {
	noteResp, err := dep.NoteFeedServer().BatchGetFeedNotes(ctx, &notev1.BatchGetFeedNotesRequest{
		NoteIds: noteIds,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz batch get note failed").WithCtx(ctx)
	}

	responses := noteResp.GetResult()
	notes := make([]*notev1.FeedNoteItem, 0, len(responses))
	for _, r := range responses {
		notes = append(notes, r.Item)
	}

	if len(notes) == 0 {
		return []*model.FeedNoteItem{}, nil
	}

	feeds, err := b.AssembleNoteFeeds(ctx, notes)
	if err != nil {
		return nil, err
	}

	feedsMap := xslice.MakeMap(feeds, func(v *model.FeedNoteItem) int64 { return int64(v.NoteId) })

	// 需要确保顺序
	ret := make([]*model.FeedNoteItem, len(noteIds))
	for idx, noteId := range noteIds {
		if n, ok := feedsMap[noteId]; ok {
			ret[idx] = n
		} else {
			ret[idx] = nil
		}
	}

	filtered := xslice.Filter(ret, func(_ int, v *model.FeedNoteItem) bool { return v == nil })

	return filtered, nil
}

func (b *Biz) GetNoteAuthor(ctx context.Context, noteId int64) (int64, error) {
	resp, err := dep.NoteFeedServer().GetNoteAuthor(ctx, &notev1.GetNoteAuthorRequest{
		NoteId: noteId,
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "remote feed server get author failed").WithCtx(ctx).WithExtras("note_id", noteId)
	}

	return resp.GetAuthor(), nil
}
