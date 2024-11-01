package note

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/safety"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	notesdk "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

var (
	IdConfuser *safety.Confuser
)

func initModel(c *config.Config) {
	IdConfuser = safety.NewConfuser(c.Metadata.Note.Salt, 24)
}

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
	NoteId string `json:"note_id"`
}

type UpdateReq struct {
	NoteId string `json:"note_id"`
	CreateReq
}

type UpdateRes struct {
	NoteId string `json:"note_id"`
}

type NoteIdReq struct {
	NoteId string `json:"note_id" path:"note_id" form:"note_id"`
}

func (r *NoteIdReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if len(r.NoteId) <= 0 {
		return xerror.ErrArgs.Msg("笔记id错误")
	}

	return nil
}

type NoteItemImage struct {
	Url  string `json:"url"`
	Type int    `json:"type"`
}

type NoteItemImageList []*NoteItemImage

type NoteItem struct {
	NoteId   string            `json:"note_id"`
	Title    string            `json:"title"`
	Desc     string            `json:"desc"`
	Privacy  int8              `json:"privacy"`
	CreateAt int64             `json:"create_at"`
	UpdateAt int64             `json:"update_at"`
	Images   NoteItemImageList `json:"images"`
	Likes    uint64            `json:"likes"`
}

func NewNoteItemFromPb(pb *notesdk.NoteItem) *NoteItem {
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

	return &NoteItem{
		NoteId:   IdConfuser.ConfuseU(pb.NoteId),
		Title:    pb.Title,
		Desc:     pb.Desc,
		Privacy:  int8(pb.Privacy),
		CreateAt: pb.CreateAt,
		UpdateAt: pb.UpdateAt,
		Images:   images,
		Likes:    pb.Likes,
	}
}

type ListRes struct {
	Items []*NoteItem `json:"items"`
}

func NewListResFromPb(pb *notesdk.ListNoteResponse) *ListRes {
	if pb == nil {
		return nil
	}

	items := make([]*NoteItem, 0, len(pb.Items))
	for _, item := range pb.Items {
		items = append(items, NewNoteItemFromPb(item))
	}

	return &ListRes{Items: items}
}

type UploadAuthReq struct {
	Resource string `json:"resource" form:"resource"`
	Source   string `json:"source" form:"source,optional"`
	MimeType string `json:"mime" form:"mime"`
}

func (r *UploadAuthReq) AsPb() *notesdk.GetUploadAuthRequest {
	return &notesdk.GetUploadAuthRequest{
		Resource: r.Resource,
		Source:   r.Source,
		MimeType: r.MimeType,
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
	NoteId string        `json:"note_id"`
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
	NoteId string `json:"note_id"`
	Count  uint64 `json:"count"`
}
