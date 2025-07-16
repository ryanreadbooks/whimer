package note

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/model"
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
	Basic  CreateReqBasic     `json:"basic"`
	Images CreateReqImageList `json:"images"`
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

	return nil
}

func (r *CreateReq) AsPb() *notev1.CreateNoteRequest {
	return &notev1.CreateNoteRequest{
		Basic:  r.Basic.AsPb(),
		Images: r.Images.AsPb(),
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
		return xerror.ErrArgs.Msg("笔记id错误")
	}

	return nil
}

type ListReq struct {
	Cursor uint64 `form:"cursor,optional"`
	Count  int32  `form:"count,optional"`
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

type NoteItemImageMeta struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
}

type NoteItemImage struct {
	Url    string            `json:"url"`
	Type   int               `json:"type"`
	Meta   NoteItemImageMeta `json:"meta"`
	UrlPrv string            `json:"url_prv"`
}

type NoteItemImageList []*NoteItemImage

// 包含发起请求的用户和该笔记的交互记录
type Interaction struct {
	Liked     bool `json:"liked"`     // 用户是否点赞过该笔记
	Commented bool `json:"commented"` // 用户是否评论过该笔记
}

type AdminNoteItem struct {
	NoteId   model.NoteId      `json:"note_id"`
	Title    string            `json:"title"`
	Desc     string            `json:"desc"`
	Privacy  int8              `json:"privacy"`
	CreateAt int64             `json:"create_at"`
	UpdateAt int64             `json:"update_at"`
	Images   NoteItemImageList `json:"images"`
	Likes    uint64            `json:"likes"`
	Replies  uint64            `json:"replies"`
	Interact Interaction       `json:"interact"`
}

func NewAdminNoteItemFromPb(pb *notev1.NoteItem) *AdminNoteItem {
	if pb == nil {
		return nil
	}

	images := make(NoteItemImageList, 0, len(pb.Images))
	for _, img := range pb.Images {
		images = append(images, &NoteItemImage{
			Url:    img.Url,
			Type:   int(img.Type),
			UrlPrv: img.UrlPrv,
			Meta: NoteItemImageMeta{
				Width:  img.Meta.Width,
				Height: img.Meta.Height,
			},
		})
	}

	return &AdminNoteItem{
		NoteId:   model.NoteId(pb.NoteId),
		Title:    pb.Title,
		Desc:     pb.Desc,
		Privacy:  int8(pb.Privacy),
		CreateAt: pb.CreateAt,
		UpdateAt: pb.UpdateAt,
		Images:   images,
		Likes:    pb.Likes,
		Replies:  pb.Replies,
	}
}

type AdminListRes struct {
	Items      []*AdminNoteItem `json:"items"`
	NextCursor uint64           `json:"next_cursor"`
	HasNext    bool             `json:"has_next"`
}

func NewListResFromPb(pb *notev1.ListNoteResponse) *AdminListRes {
	if pb == nil {
		return nil
	}

	items := make([]*AdminNoteItem, 0, len(pb.Items))
	for _, item := range pb.Items {
		items = append(items, NewAdminNoteItemFromPb(item))
	}

	return &AdminListRes{
		Items:      items,
		NextCursor: pb.NextCursor,
		HasNext:    pb.HasNext,
	}
}

type AdminPageListRes struct {
	Items []*AdminNoteItem `json:"items"`
	Total uint64           `json:"total"`
}

func NewPageListResFromPb(pb *notev1.PageListNoteResponse) *AdminPageListRes {
	if pb == nil {
		return nil
	}

	items := make([]*AdminNoteItem, 0, len(pb.Items))
	for _, item := range pb.Items {
		items = append(items, NewAdminNoteItemFromPb(item))
	}

	return &AdminPageListRes{
		Items: items,
		Total: uint64(pb.Total),
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

type LikeReqAction uint8

const (
	LikeReqActionUndo LikeReqAction = 0
	LikeReqActionDo   LikeReqAction = 1
)

// 点赞/取消点赞
type LikeReq struct {
	NoteId uint64        `json:"note_id"`
	Action LikeReqAction `json:"action"`
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
	NoteId uint64 `json:"note_id"`
	Count  uint64 `json:"count"`
}
