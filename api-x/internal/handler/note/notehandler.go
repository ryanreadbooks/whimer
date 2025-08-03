package note

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/api-x/internal/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/model/errors"
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"

	"github.com/zeromicro/go-zero/rest/httpx"
	"golang.org/x/sync/errgroup"
)

type Handler struct{}

func NewHandler(c *config.Config) *Handler {
	return &Handler{}
}

func (h *Handler) hasNoteCheck(ctx context.Context, noteId int64) error {
	if resp, err := infra.NoteCreatorServer().IsNoteExist(ctx,
		&notev1.IsNoteExistRequest{
			NoteId: noteId,
		}); err != nil {
		return err
	} else {
		if !resp.Exist {
			return errors.ErrNoteNotFound
		}
	}

	return nil
}

func (h *Handler) CreatorCreateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CreateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		// service to create note
		resp, err := infra.NoteCreatorServer().CreateNote(r.Context(), req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, CreateRes{NoteId: model.NoteId(resp.NoteId)})
	}
}

func (h *Handler) CreatorUpdateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UpdateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		_, err = infra.NoteCreatorServer().UpdateNote(r.Context(), &notev1.UpdateNoteRequest{
			NoteId: int64(req.NoteId),
			Note: &notev1.CreateNoteRequest{
				Basic:  req.Basic.AsPb(),
				Images: req.Images.AsPb(),
			},
		})

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, UpdateRes{NoteId: req.NoteId})
	}
}

func (h *Handler) CreatorDeleteNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		_, err = infra.NoteCreatorServer().DeleteNote(r.Context(), &notev1.DeleteNoteRequest{
			NoteId: int64(req.NoteId),
		})

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) assignNoteExtra(ctx context.Context, notes []*model.AdminNoteItem) {
	var (
		noteIds      = make([]int64, 0, len(notes))
		oidLiked     = make(map[int64]bool)
		oidCommented = make(map[int64]bool)
		uid          = metadata.Uid(ctx)
		eg           errgroup.Group
	)

	for _, n := range notes {
		noteIds = append(noteIds, int64(n.NoteId))
	}

	eg.Go(func() error {
		mappings := make(map[int64]*notev1.NoteIdList)
		mappings[uid] = &notev1.NoteIdList{
			NoteIds: noteIds,
		}

		// 点赞信息
		resp, err := infra.NoteInteractServer().BatchCheckUserLikeStatus(ctx,
			&notev1.BatchCheckUserLikeStatusRequest{
				Mappings: mappings,
			})
		if err != nil {
			return xerror.Wrapf(err, "failed to get user like status").WithCtx(ctx)
		}

		pairs := resp.GetResults()
		for _, likedInfo := range pairs[uid].GetList() {
			oidLiked[likedInfo.NoteId] = likedInfo.Liked
		}

		for _, note := range notes {
			noteId := int64(note.NoteId)
			note.Interact.Liked = oidLiked[noteId]
		}

		return nil
	})

	eg.Go(func() error {
		commentMappings := make(map[int64]*commentv1.BatchCheckUserOnObjectRequest_Objects)
		commentMappings[uid] = &commentv1.BatchCheckUserOnObjectRequest_Objects{
			Oids: noteIds,
		}
		// 评论信息
		resp, err := infra.Commenter().BatchCheckUserOnObject(ctx,
			&commentv1.BatchCheckUserOnObjectRequest{
				Mappings: commentMappings,
			})
		if err != nil {
			return xerror.Wrapf(err, "failed to get comment status").WithCtx(ctx)
		}

		// organize result
		pairs := resp.GetResults()
		for _, comInfo := range pairs[uid].GetList() {
			oidCommented[comInfo.Oid] = comInfo.Commented
		}
		for _, note := range notes {
			noteId := int64(note.NoteId)
			note.Interact.Commented = oidCommented[noteId]
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		xlog.Msgf("failed to assign note extra").Err(err).Errorx(ctx)
		return
	}

	for _, note := range notes {
		noteId := int64(note.NoteId)
		note.Interact.Liked = oidLiked[noteId]
		note.Interact.Commented = oidCommented[noteId]
	}
}

func (h *Handler) CreatorListNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[ListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		resp, err := infra.NoteCreatorServer().ListNote(ctx, &notev1.ListNoteRequest{
			Cursor: req.Cursor,
			Count:  req.Count,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}
		result := NewListResFromPb(resp)
		h.assignNoteExtra(ctx, result.Items)

		xhttp.OkJson(w, result)
	}
}

func (h *Handler) CreatorPageListNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[PageListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := infra.NoteCreatorServer().PageListNote(ctx, &notev1.PageListNoteRequest{
			Page:  req.Page,
			Count: req.Count,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		result := NewPageListResFromPb(resp)
		h.assignNoteExtra(ctx, result.Items)

		xhttp.OkJson(w, result)
	}
}

func (h *Handler) CreatorGetNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		resp, err := infra.NoteCreatorServer().GetNote(ctx, &notev1.GetNoteRequest{
			NoteId: int64(req.NoteId),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		result := model.NewAdminNoteItemFromPb(resp.Note)
		h.assignNoteExtra(ctx, []*model.AdminNoteItem{result})
		xhttp.OkJson(w, result)
	}
}

func (h *Handler) CreatorUploadNoteAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := infra.NoteCreatorServer().BatchGetUploadAuth(r.Context(), req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *Handler) CreatorUploadNoteAuthV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := infra.NoteCreatorServer().BatchGetUploadAuthV2(r.Context(), req.AsPbV2())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 点赞/取消点赞笔记
func (h *Handler) LikeNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			uid = metadata.Uid(r.Context())
		)

		req, err := xhttp.ParseValidate[LikeReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		nid := req.NoteId
		_, err = infra.NoteInteractServer().LikeNote(r.Context(), &notev1.LikeNoteRequest{
			NoteId:    int64(nid),
			Uid:       uid,
			Operation: notev1.LikeNoteRequest_Operation(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}
		xhttp.OkJson(w, nil)
	}
}

// 获取笔记点赞数量
func (h *Handler) GetNoteLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		nid := int64(req.NoteId)
		resp, err := infra.NoteInteractServer().GetNoteLikes(r.Context(),
			&notev1.GetNoteLikesRequest{NoteId: nid})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &GetLikesRes{
			Count:  resp.Likes,
			NoteId: resp.NoteId,
		})
	}
}

// TODO 获取点赞过的笔记
func (h *Handler) GetLikeNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// 创建新标签
func (h *Handler) AddNewTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[AddTagReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := infra.NoteCreatorServer().AddTag(ctx, &notev1.AddTagRequest{
			Name: req.Name,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &AddTagRes{TagId: model.TagId(resp.Id)})
	}
}
