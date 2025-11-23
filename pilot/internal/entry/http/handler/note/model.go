package note

import (
	"strings"
	"unicode"
	"unicode/utf8"

	feedmodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"

	"github.com/ryanreadbooks/whimer/misc/xerror"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

const (
	VisibilityPublic  = int8(notev1.NotePrivacy_PUBLIC)
	VisibilityPrivate = int8(notev1.NotePrivacy_PRIVATE)
)

const (
	AssetTypeImage = int8(notev1.NoteAssetType_IMAGE)
	AssetTypeVideo = int8(notev1.NoteAssetType_VIDEO)
)

const (
	maxTitleLen   = 48
	maxDescLen    = 2048
	maxTagNameLen = 255
)

type CreateReqBasic struct {
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Privacy int8   `json:"privacy"`
}

func (v *CreateReqBasic) Validate() error {
	if v.Privacy != VisibilityPublic && v.Privacy != VisibilityPrivate {
		return xerror.ErrArgs.Msg("不支的可见范围")
	}

	titleLen := utf8.RuneCountInString(v.Title)
	if titleLen > maxTitleLen {
		return xerror.ErrArgs.Msg("标题长度错误")
	}
	descLen := utf8.RuneCountInString(v.Desc)
	if descLen > maxDescLen {
		return xerror.ErrArgs.Msg("简介超长")
	}

	return nil
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
	TagList []TagId            `json:"tag_list,omitempty,optional"`
	AtUsers model.AtUserList   `json:"at_users,omitempty,optional"`
}

type TagId struct { // 必须再包一层 直接用数组无法解析
	Id model.TagId `json:"id"`
}

func (r *CreateReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if err := r.Basic.Validate(); err != nil {
		return err
	}

	for _, img := range r.Images {
		if img.FileId == "" {
			return xerror.ErrArgs.Msg("上传图片无名")
		}

		if img.Width == 0 || img.Height == 0 || img.Format == "" {
			return xerror.ErrArgs.Msg("上传图片未指定图片信息")
		}

		if err := model.CheckImageFormat(img.Format); err != nil {
			return err
		}
	}

	if len(r.TagList) > 10 {
		return xerror.ErrArgs.Msg("标签超出限制")
	}

	r.AtUsers = r.AtUsers.Filter()

	return nil
}

func AtUsersAsPb(atUsers model.AtUserList) []*notev1.NoteAtUser {
	users := make([]*notev1.NoteAtUser, 0, len(atUsers))
	for _, u := range atUsers {
		users = append(users, &notev1.NoteAtUser{
			Nickname: u.Nickname,
			Uid:      u.Uid,
		})
	}

	return users
}

func (r *CreateReq) AsPb() *notev1.CreateNoteRequest {
	tagList := []int64{}
	for _, t := range r.TagList {
		tagList = append(tagList, int64(t.Id))
	}
	return &notev1.CreateNoteRequest{
		Basic:   r.Basic.AsPb(),
		Images:  r.Images.AsPb(),
		Tags:    &notev1.CreateReqTag{TagList: tagList},
		AtUsers: AtUsersAsPb(r.AtUsers),
	}
}

type CreateRes struct {
	NoteId model.NoteId `json:"note_id"`
}

type UpdateReq struct {
	NoteId model.NoteId `json:"note_id"`
	CreateReq
}

func (r *UpdateReq) Validate() error {
	if r.NoteId <= 0 {
		return errors.ErrNoteNotFound
	}

	if err := r.CreateReq.Validate(); err != nil {
		return err
	}

	return nil
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

const (
	maxCountOfUploadAuth = 8
)

func (r *UploadAuthReq) Validate() error {
	if r.Count <= 0 {
		r.Count = 1
	}

	if r.Count > maxCountOfUploadAuth {
		return xerror.ErrInvalidArgs.Msg("不支持请求这么多上传凭证")
	}

	return nil
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
	NoteId model.NoteId `json:"note_id"`
	Count  int64        `json:"count"`
}

func checkTagName(s string) error {
	if !utf8.ValidString(s) {
		return xerror.ErrInvalidArgs.Msg("不支持的字符格式")
	}

	for _, r := range s {
		if unicode.IsSpace(r) {
			return xerror.ErrInvalidArgs.Msg("标签不能存在空格")
		}
	}

	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return xerror.ErrInvalidArgs.Msg("标签名不能包含特殊字符")
		}
	}

	return nil
}

// 新增标签
type AddTagReq struct {
	Name string `json:"name"`
}

func (r *AddTagReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.Name == "" {
		return xerror.ErrInvalidArgs.Msg("标签名为空")
	}

	if l := utf8.RuneCountInString(r.Name); l > maxTagNameLen {
		return xerror.ErrInvalidArgs.Msg("标签名太长")
	}

	if err := checkTagName(r.Name); err != nil {
		return err
	}

	return nil
}

type AddTagRes struct {
	TagId model.TagId `json:"tag_id"`
}

type SearchTagsReq struct {
	Name string `json:"name"`
}

func (r *SearchTagsReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return xerror.ErrInvalidArgs.Msg("输入标签目标")
	}

	if err := checkTagName(r.Name); err != nil {
		return err
	}

	return nil
}

type SearchedNoteTag struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type SearchTagsRes struct {
	Items []SearchedNoteTag `json:"items"`
}

type GetLikedNoteRequest struct {
	Uid    int64  `form:"uid"`
	Cursor string `form:"cursor,optional"`
	Count  int32  `form:"count,optional"`
}

func (r *GetLikedNoteRequest) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	r.Count = min(r.Count, 20)
	if r.Count <= 0 {
		r.Count = 10
	}

	return nil
}

type GetLikedNoteResponse struct {
	Items      []*feedmodel.FeedNoteItem `json:"items"`
	NextCursor string                    `json:"next_cursor"`
	HasNext    bool                      `json:"has_next"`
}
