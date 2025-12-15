package model

import (
	"fmt"
	"path/filepath"
	"strings"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type NoteImageMeta struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
}

// 笔记图片资源
type NoteImage struct {
	Key  string        `json:"url"`
	Type int           `json:"type"`
	Meta NoteImageMeta `json:"meta"`

	bucket string `json:"-"` // 非全场景必须字段 用到时手动Set
}

func (i *NoteImage) SetBucket(bucket string) {
	if i == nil {
		return
	}
	i.bucket = bucket
}

func (i *NoteImage) GetBucket() string {
	if i == nil {
		return ""
	}
	return i.bucket
}

type NoteImageList []*NoteImage

func (l NoteImageList) AsPb() []*notev1.NoteImage {
	images := make([]*notev1.NoteImage, 0, len(l))
	for _, img := range l {
		images = append(images, &notev1.NoteImage{
			Key:  img.Key,
			Type: int32(img.Type),
			Meta: &notev1.NoteImageMeta{
				Width:  img.Meta.Width,
				Height: img.Meta.Height,
				Format: img.Meta.Format,
			},
		})
	}
	return images
}

func (l NoteImageList) SetBucket(bucket string) {
	for _, img := range l {
		img.SetBucket(bucket)
	}
}

type NoteVideoMedia struct {
	Width        int32  `json:"width"`
	Height       int32  `json:"height"`
	VideoCodec   string `json:"video_codec"`
	VideoBitrate int32  `json:"video_bitrate"`
	FrameRate    int32  `json:"frame_rate"`
	Duration     int32  `json:"duration"`
	Format       string `json:"format"`
	AudioCodec   string `json:"audio_codec"`
	AudioBitrate int32  `json:"audio_bitrate"`
}

type NoteVideoItem struct {
	Key   string          `json:"key"` // 资源key 包含bucket名
	Media *NoteVideoMedia `json:"meta"`

	bucket string `json:"-"` // 非全场景必须字段 用到时手动Set
}

func (v *NoteVideoItem) SetBucket(bucket string) {
	if v == nil {
		return
	}
	v.bucket = bucket
}

func (v *NoteVideoItem) GetBucket() string {
	if v == nil {
		return ""
	}
	return v.bucket
}

func (v *NoteVideoItem) TrimBucket() string {
	if v == nil {
		return ""
	}

	if strings.HasPrefix(v.Key, v.GetBucket()+"/") {
		return strings.TrimPrefix(v.Key, v.GetBucket()+"/")
	}

	return v.Key
}

type SupportedVideoSuffix string

const (
	SupportedVideoH264Suffix SupportedVideoSuffix = "_264"
	SupportedVideoH265Suffix SupportedVideoSuffix = "_265"
	SupportedVideoAV1Suffix  SupportedVideoSuffix = "_av1"
)

// 笔记视频资源
type NoteVideo struct {
	H264 *NoteVideoItem `json:"h264,omitempty"`
	H265 *NoteVideoItem `json:"h265,omitempty"`
	AV1  *NoteVideoItem `json:"av1,omitempty"`

	rawUrl    string `json:"-"` // 非必要字段 需要时填充
	rawBucket string `json:"-"` // 非必要字段 需要时填充
}

func (v *NoteVideo) GetRawUrl() string {
	if v == nil {
		return ""
	}
	return v.rawUrl
}

func (v *NoteVideo) SetRawUrl(url string) {
	if v == nil {
		return
	}
	v.rawUrl = url
}

func (v *NoteVideo) SetRawBucket(bucket string) {
	if v == nil {
		return
	}
	v.rawBucket = bucket
}

func (v *NoteVideo) GetRawBucket() string {
	if v == nil {
		return ""
	}
	return v.rawBucket
}

func (v *NoteVideo) SetTargetBucket(bucket string) {
	if v == nil {
		return
	}
	v.H264.SetBucket(bucket)
	v.H265.SetBucket(bucket)
	v.AV1.SetBucket(bucket)
}

func FormatNoteVideoKey(id string, suffix SupportedVideoSuffix) string {
	ext := filepath.Ext(id)
	basename := id
	if ext != "" {
		basename = strings.TrimSuffix(id, ext)
	}

	return fmt.Sprintf("%s%s.mp4", basename, suffix)
}

type Note struct {
	NoteId   int64         `json:"note_id"`
	Title    string        `json:"title"`
	Desc     string        `json:"desc"`
	Privacy  Privacy       `json:"privacy,omitempty"`
	Type     NoteType      `json:"type"`
	State    NoteState     `json:"state"` // 笔记状态
	CreateAt int64         `json:"create_at,omitempty"`
	UpdateAt int64         `json:"update_at,omitempty"`
	Ip       string        `json:"ip"`
	Images   NoteImageList `json:"images"`
	Videos   *NoteVideo    `json:"videos"`  // 包含多种编码的视频资源
	Likes    int64         `json:"likes"`   // 点赞数
	Replies  int64         `json:"replies"` // 评论数

	// ext字段
	Tags    []*NoteTag `json:"tags,omitempty"`
	AtUsers []*AtUser  `json:"at_users,omitempty"`

	// unexported to user
	Owner int64 `json:"-"`
}

func (n *Note) AsSlice() []*Note {
	return []*Note{n}
}

func (i *Note) AsPb() *notev1.NoteItem {
	res := &notev1.NoteItem{
		NoteId:   i.NoteId,
		Title:    i.Title,
		Desc:     i.Desc,
		Privacy:  int32(i.Privacy),
		State:    notev1.NoteState(i.State),
		NoteType: notev1.NoteAssetType(i.Type),
		CreateAt: i.CreateAt,
		UpdateAt: i.UpdateAt,
		Ip:       i.Ip,
		Images:   i.Images.AsPb(),
		Likes:    i.Likes,
		Replies:  i.Replies,
		Owner:    i.Owner,
	}

	// note tags
	res.Tags = NoteTagListAsPb(i.Tags)
	// at_users
	res.AtUsers = AtUsersAsPb(i.AtUsers)

	return res
}

// 转换成pb并隐藏一些不公开的属性
func (i *Note) AsFeedPb() *notev1.FeedNoteItem {
	return &notev1.FeedNoteItem{
		NoteId:    i.NoteId,
		Title:     i.Title,
		Desc:      i.Desc,
		NoteType:  notev1.NoteAssetType(i.Type),
		CreatedAt: i.CreateAt,
		UpdatedAt: i.UpdateAt,
		Images:    i.Images.AsPb(),
		Ip:        i.Ip,
		Likes:     i.Likes,
		Author:    i.Owner,
		Replies:   i.Replies,
	}
}

type Notes struct {
	Items []*Note `json:"items"`
}

func (n *Notes) GetIds() []int64 {
	r := make([]int64, 0, len(n.Items))
	for _, item := range n.Items {
		r = append(r, item.NoteId)
	}
	return r
}

func PbFeedNoteItemsFromNotes(ns *Notes) []*notev1.FeedNoteItem {
	items := make([]*notev1.FeedNoteItem, 0, len(ns.Items))
	for _, item := range ns.Items {
		items = append(items, item.AsFeedPb())
	}

	return items
}

// 每个用户和笔记的交互情况
type UserInteraction struct {
	Liked bool // 是否点赞
}

func (u *UserInteraction) AsPb() *notev1.NoteInteraction {
	return &notev1.NoteInteraction{
		Liked: u.Liked,
	}
}

type LikeStatus struct {
	NoteId int64
	Liked  bool
}

type InteractStatus struct {
	NoteId    int64
	Liked     bool
	Starred   bool
	Commented bool
}
