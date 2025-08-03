package note

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/model/errors"

	"github.com/ryanreadbooks/whimer/misc/xerror"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type CreateReqBasic struct {
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Privacy int    `json:"privacy"`
}

func (b *CreateReqBasic) AsPb() *notev1.CreateReqBasic {
	return &notev1.CreateReqBasic{
		Title:   b.Title,
		Desc:    b.Desc,
		Privacy: int32(b.Privacy),
	}
}

type CreateReqImage struct {
	FileId string `json:"file_id"`
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
}

type CreateReqImageList []CreateReqImage

func (r CreateReqImageList) AsPb() []*notev1.CreateReqImage {
	images := make([]*notev1.CreateReqImage, 0, len(r))
	for _, img := range r {
		images = append(images, &notev1.CreateReqImage{
			FileId: img.FileId,
			Width:  img.Width,
			Height: img.Height,
			Format: img.Format,
		})
	}

	return images
}

type CreateReq struct {
	Basic   CreateReqBasic     `json:"basic"`
	Images  CreateReqImageList `json:"images"`
	TagList []struct {         // 必须再包一层 直接用数组无法解析
		Id model.TagId `json:"id"`
	} `json:"tag_list,omitempty"`
}

func (r *CreateReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	for _, img := range r.Images {
		if img.FileId == "" {
			return xerror.ErrArgs.Msg("上传图片无名")
		}

		if img.Width == 0 || img.Height == 0 || img.Format == "" {
			return xerror.ErrArgs.Msg("上传图片未指定图片信息")
		}
	}

	if len(r.TagList) > 10 {
		return xerror.ErrArgs.Msg("标签超出限制")
	}

	return nil
}

func (r *CreateReq) AsPb() *notev1.CreateNoteRequest {
	tagList := []int64{}
	for _, t := range r.TagList {
		tagList = append(tagList, int64(t.Id))
	}
	return &notev1.CreateNoteRequest{
		Basic:  r.Basic.AsPb(),
		Images: r.Images.AsPb(),
		Tags:   &notev1.CreateReqTag{TagList: tagList},
	}
}

type CreateRes struct {
	NoteId model.NoteId `json:"note_id"`
}

type UpdateReq struct {
	NoteId model.NoteId `json:"note_id"`
	CreateReq
}

type UpdateRes struct {
	NoteId model.NoteId `json:"note_id"`
}

type NoteIdReq struct {
	NoteId model.NoteId `json:"note_id" path:"note_id" form:"note_id"`
}

func (r *NoteIdReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.NoteId <= 0 {
		return errors.ErrNoteNotFound
	}

	return nil
}

type ListReq struct {
	Cursor int64 `form:"cursor,optional"`
	Count  int32 `form:"count,optional"`
}

func (r *ListReq) Validate() error {
	if r.Count == 0 {
		r.Count = 15
	}
	if r.Count >= 15 {
		r.Count = 15
	}

	return nil
}

type PageListReq struct {
	Page  int32 `form:"page,optional"`
	Count int32 `form:"count,default=15"`
}

func (r *PageListReq) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Count >= 15 {
		r.Count = 15
	}

	return nil
}

type AdminListRes struct {
	Items      []*model.AdminNoteItem `json:"items"`
	NextCursor int64                  `json:"next_cursor"`
	HasNext    bool                   `json:"has_next"`
}

func NewListResFromPb(pb *notev1.ListNoteResponse) *AdminListRes {
	if pb == nil {
		return nil
	}

	items := make([]*model.AdminNoteItem, 0, len(pb.Items))
	for _, item := range pb.Items {
		items = append(items, model.NewAdminNoteItemFromPb(item))
	}

	return &AdminListRes{
		Items:      items,
		NextCursor: pb.NextCursor,
		HasNext:    pb.HasNext,
	}
}

type AdminPageListRes struct {
	Items []*model.AdminNoteItem `json:"items"`
	Total int64                  `json:"total"`
}

func NewPageListResFromPb(pb *notev1.PageListNoteResponse) *AdminPageListRes {
	if pb == nil {
		return nil
	}

	items := make([]*model.AdminNoteItem, 0, len(pb.Items))
	for _, item := range pb.Items {
		items = append(items, model.NewAdminNoteItemFromPb(item))
	}

	return &AdminPageListRes{
		Items: items,
		Total: int64(pb.Total),
	}
}

type UploadAuthReq struct {
	Resource string `form:"resource"`
	Source   string `form:"source,optional"`
	Count    int32  `form:"count,optional"`
}

func (r *UploadAuthReq) Validate() error {
	if r.Count <= 0 {
		r.Count = 1
	}

	if r.Count > 8 {
		return xerror.ErrInvalidArgs.Msg("不支持请求这么多上传凭证")
	}

	return nil
}

func (r *UploadAuthReq) AsPb() *notev1.BatchGetUploadAuthRequest {
	return &notev1.BatchGetUploadAuthRequest{
		Resource: r.Resource,
		Source:   r.Source,
		Count:    r.Count,
	}
}

func (r *UploadAuthReq) AsPbV2() *notev1.BatchGetUploadAuthV2Request {
	return &notev1.BatchGetUploadAuthV2Request{
		Resource: r.Resource,
		Source:   r.Source,
		Count:    r.Count,
	}
}

type UploadAuthResHeaders struct {
	Auth   string `json:"auth"`
	Sha256 string `json:"sha256"`
	Date   string `json:"date"`
	Token  string `json:"token"`
}

// 上传凭证响应
type UploadAuthRes struct {
	FildId      string               `json:"fild_id"`
	CurrentTime int64                `json:"current_time"`
	ExpireTime  int64                `json:"expire_time"`
	UploadAddr  string               `json:"upload_addr"`
	Headers     UploadAuthResHeaders `json:"headers"`
}

// 点赞/取消点赞
type LikeReq struct {
	NoteId model.NoteId        `json:"note_id"`
	Action model.LikeReqAction `json:"action"`
}

func (r *LikeReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.Action != 0 && r.Action != 1 {
		return xerror.ErrInvalidArgs.Msg("不支持的点赞操作")
	}

	return nil
}

type GetLikesRes struct {
	NoteId int64 `json:"note_id"`
	Count  int64 `json:"count"`
}
