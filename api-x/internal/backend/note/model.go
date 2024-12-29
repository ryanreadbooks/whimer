package note

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	notesdk "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

type CreateReqBasic struct {
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Privacy int    `json:"privacy"`
}

func (b *CreateReqBasic) AsPb() *notesdk.CreateReqBasic {
	return &notesdk.CreateReqBasic{
		Title:   b.Title,
		Desc:    b.Desc,
		Privacy: int32(b.Privacy),
	}
}

type CreateReqImage struct {
	FileId string `json:"file_id"`
}

type CreateReqImageList []CreateReqImage

func (r CreateReqImageList) AsPb() []*notesdk.CreateReqImage {
	images := make([]*notesdk.CreateReqImage, 0, len(r))
	for _, img := range r {
		images = append(images, &notesdk.CreateReqImage{FileId: img.FileId})
	}

	return images
}

type CreateReq struct {
	Basic  CreateReqBasic     `json:"basic"`
	Images CreateReqImageList `json:"images"`
}

func (r *CreateReq) AsPb() *notesdk.CreateNoteRequest {
	return &notesdk.CreateNoteRequest{
		Basic:  r.Basic.AsPb(),
		Images: r.Images.AsPb(),
	}
}

type CreateRes struct {
	NoteId uint64 `json:"note_id"`
}

type UpdateReq struct {
	NoteId uint64 `json:"note_id"`
	CreateReq
}

type UpdateRes struct {
	NoteId uint64 `json:"note_id"`
}

type NoteIdReq struct {
	NoteId uint64 `json:"note_id" path:"note_id" form:"note_id"`
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

type NoteItemImage struct {
	Url  string `json:"url"`
	Type int    `json:"type"`
}

type NoteItemImageList []*NoteItemImage

type AdminNoteItem struct {
	NoteId   uint64            `json:"note_id"`
	Title    string            `json:"title"`
	Desc     string            `json:"desc"`
	Privacy  int8              `json:"privacy"`
	CreateAt int64             `json:"create_at"`
	UpdateAt int64             `json:"update_at"`
	Images   NoteItemImageList `json:"images"`
	Likes    uint64            `json:"likes"`
}

func NewAdminNoteItemFromPb(pb *notesdk.NoteItem) *AdminNoteItem {
	if pb == nil {
		return nil
	}

	images := make(NoteItemImageList, 0, len(pb.Images))
	for _, img := range pb.Images {
		images = append(images, &NoteItemImage{
			Url:  img.Url,
			Type: int(img.Type),
		})
	}

	return &AdminNoteItem{
		NoteId:   pb.NoteId,
		Title:    pb.Title,
		Desc:     pb.Desc,
		Privacy:  int8(pb.Privacy),
		CreateAt: pb.CreateAt,
		UpdateAt: pb.UpdateAt,
		Images:   images,
		Likes:    pb.Likes,
	}
}

type AdminListRes struct {
	Items      []*AdminNoteItem `json:"items"`
	NextCursor uint64           `json:"next_cursor"`
	HasNext    bool             `json:"has_next"`
}

func NewListResFromPb(pb *notesdk.ListNoteResponse) *AdminListRes {
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

func (r *UploadAuthReq) AsPb() *notesdk.BatchGetUploadAuthRequest {
	return &notesdk.BatchGetUploadAuthRequest{
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
