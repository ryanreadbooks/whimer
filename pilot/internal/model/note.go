package model

import (
	"context"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"

	"github.com/ryanreadbooks/whimer/misc/imgproxy"
)

type NoteType string

// 笔记类型
const (
	NoteTypeImage = "image"
	NoteTypeVideo = "video"
)

var (
	noteTypePbMapping = map[NoteType]notev1.NoteAssetType{
		NoteTypeImage: notev1.NoteAssetType_IMAGE,
		NoteTypeVideo: notev1.NoteAssetType_VIDEO,
	}

	noteTypePbReverseMapping = map[notev1.NoteAssetType]NoteType{
		notev1.NoteAssetType_IMAGE: NoteTypeImage,
		notev1.NoteAssetType_VIDEO: NoteTypeVideo,
	}
)

func IsNoteTypeValid(t NoteType) bool {
	_, ok := noteTypePbMapping[t]
	return ok
}

func ConvertNoteTypeAsPb(n NoteType) notev1.NoteAssetType {
	if t, ok := noteTypePbMapping[n]; ok {
		return t
	}
	return notev1.NoteAssetType_NOTE_ASSET_TYPE_UNSPECIFIED
}

func ConvertNoteTypeFromPb(n notev1.NoteAssetType) NoteType {
	if t, ok := noteTypePbReverseMapping[n]; ok {
		return t
	}

	return ""
}

type NoteItemImageMeta struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
}

type NoteItemImage struct {
	Key      string            `json:"key,omitempty"`
	Url      string            `json:"url"`
	Type     int               `json:"type"`
	Metadata NoteItemImageMeta `json:"metadata"`
	UrlPrv   string            `json:"url_prv"`
}

type NoteItemImageList []*NoteItemImage

type NoteItemVideo struct {
	Key      string            `json:"-"`
	Url      string            `json:"url"`
	Type     int               `json:"type"`
	Metadata NoteItemVideoMeta `json:"metadata"`
}

type NoteItemVideoMeta struct {
	Width      uint32  `json:"width"`
	Height     uint32  `json:"height"`
	Format     string  `json:"format"`
	Duration   float64 `json:"duration"`
	Bitrate    int64   `json:"bitrate"`
	Codec      string  `json:"codec"`
	Framerate  float64 `json:"framerate"`
	AudioCodec string  `json:"audio_codec"`
}

type NoteItemVideoList []*NoteItemVideo

// 包含发起请求的用户和该笔记的交互记录
type Interaction struct {
	Liked     bool `json:"liked"`     // 用户是否点赞过该笔记
	Commented bool `json:"commented"` // 用户是否评论过该笔记
}

type NoteTag struct {
	Id   TagId  `json:"id"`
	Name string `json:"name"`
}

func NoteTagFromPb(t *notev1.NoteTag) *NoteTag {
	return &NoteTag{
		Id:   TagId(t.Id),
		Name: t.Name,
	}
}

func NoteTagsFromPbs(ts []*notev1.NoteTag) []*NoteTag {
	if len(ts) == 0 {
		return []*NoteTag{}
	}

	var r = make([]*NoteTag, 0, len(ts))
	for _, t := range ts {
		r = append(r, NoteTagFromPb(t))
	}
	return r
}

func AtUsersFromNotePbs(us []*notev1.NoteAtUser) []*AtUser {
	if len(us) == 0 {
		return []*AtUser{}
	}

	var r = make([]*AtUser, 0, len(us))
	for _, u := range us {
		r = append(r, &AtUser{
			Nickname: u.Nickname,
			Uid:      u.Uid,
		})
	}
	return r
}

type AdminNoteItem struct {
	NoteId   NoteId            `json:"note_id"`
	Title    string            `json:"title"`
	Desc     string            `json:"desc"`
	Privacy  int8              `json:"privacy"`
	CreateAt int64             `json:"create_at"`
	UpdateAt int64             `json:"update_at"`
	Ip       string            `json:"-"`
	IpLoc    string            `json:"ip_loc"`
	Type     NoteType          `json:"type"`
	Images   NoteItemImageList `json:"images,omitempty"`
	Videos   NoteItemVideoList `json:"videos,omitempty"`
	Likes    int64             `json:"likes"`
	Replies  int64             `json:"replies"`
	Interact Interaction       `json:"interact"`
	TagList  []*NoteTag        `json:"tag_list,omitempty"`
	AtUsers  []*AtUser         `json:"at_users,omitempty"`
}

func NewNoteImageItemUrl(pbimg *notev1.NoteImage) string {
	noteAssetBucket := config.Conf.UploadResourceDefineMap["note_image"].Bucket

	url := imgproxy.GetSignedUrl(
		config.Conf.Oss.DisplayEndpointBucket(noteAssetBucket),
		pbimg.Key,
		config.Conf.ImgProxyAuth.GetKey(),
		config.Conf.ImgProxyAuth.GetSalt(),
		imgproxy.WithQuality("15"))
	return url
}

func NewNoteImageItemUrlPrv(pbimg *notev1.NoteImage) string {
	noteAssetBucket := config.Conf.UploadResourceDefineMap["note_image"].Bucket

	url := imgproxy.GetSignedUrl(
		config.Conf.Oss.DisplayEndpointBucket(noteAssetBucket),
		pbimg.GetKey(),
		config.Conf.ImgProxyAuth.GetKey(),
		config.Conf.ImgProxyAuth.GetSalt(),
		imgproxy.WithQuality("1"))
	return url
}

func NewNoteImageFromPb(pbimg *notev1.NoteImage, showKey bool) *NoteItemImage {
	url := NewNoteImageItemUrl(pbimg)
	urlPrv := NewNoteImageItemUrlPrv(pbimg)

	img := &NoteItemImage{
		Url:    url,
		Type:   int(pbimg.GetType()),
		UrlPrv: urlPrv,
		Metadata: NoteItemImageMeta{
			Width:  pbimg.GetMeta().GetWidth(),
			Height: pbimg.GetMeta().GetHeight(),
			Format: pbimg.GetMeta().GetFormat(),
		},
	}
	if showKey {
		img.Key = pbimg.Key
	}
	return img
}

func NewNoteVideoUrl(pbvideo *notev1.NoteVideo) string {
	// 应该要用s3预签名url
	url, err := dep.PresignGetUrl(context.Background(), uploadresource.NoteVideo, pbvideo.GetKey())
	if err != nil {
		return ""
	}

	return url
}

func NewNoteVideoFromPb(pbvideo *notev1.NoteVideo) *NoteItemVideo {
	if pbvideo == nil {
		return nil
	}

	url := NewNoteVideoUrl(pbvideo)
	return &NoteItemVideo{
		Key:  pbvideo.GetKey(),
		Url:  url,
		Type: int(pbvideo.GetType()),
		Metadata: NoteItemVideoMeta{
			Width:      pbvideo.GetMeta().GetWidth(),
			Height:     pbvideo.GetMeta().GetHeight(),
			Format:     pbvideo.GetMeta().GetFormat(),
			Duration:   pbvideo.GetMeta().GetDuration(),
			Bitrate:    pbvideo.GetMeta().GetBitrate(),
			Codec:      pbvideo.GetMeta().GetCodec(),
			Framerate:  pbvideo.GetMeta().GetFramerate(),
			AudioCodec: pbvideo.GetMeta().GetAudioCodec(),
		},
	}
}

func NewAdminNoteItemFromPb(pb *notev1.NoteItem) *AdminNoteItem {
	if pb == nil {
		return nil
	}

	var (
		images NoteItemImageList = make(NoteItemImageList, 0)
		videos NoteItemVideoList
	)

	if len(pb.Images) > 0 {
		images = make(NoteItemImageList, 0, len(pb.Images))
		for _, pbimg := range pb.Images {
			images = append(images, NewNoteImageFromPb(pbimg, true))
		}
	}

	if len(pb.Videos) > 0 {
		videos = make(NoteItemVideoList, 0, len(pb.Videos))
		for _, pbvideo := range pb.Videos {
			videos = append(videos, NewNoteVideoFromPb(pbvideo))
		}
	}

	var tagList []*NoteTag = NoteTagsFromPbs(pb.GetTags())
	var atUsers []*AtUser = AtUsersFromNotePbs(pb.GetAtUsers())

	ctx := context.Background()
	ipLoc, _ := infra.Ip2Loc().Convert(ctx, pb.Ip)
	return &AdminNoteItem{
		NoteId:   NoteId(pb.NoteId),
		Title:    pb.Title,
		Desc:     pb.Desc,
		Privacy:  int8(pb.Privacy),
		Type:     ConvertNoteTypeFromPb(pb.NoteType),
		CreateAt: pb.CreateAt,
		UpdateAt: pb.UpdateAt,
		Images:   images,
		Videos:   videos,
		Likes:    pb.Likes,
		Replies:  pb.Replies,
		TagList:  tagList,
		Ip:       pb.Ip,
		IpLoc:    ipLoc,
		AtUsers:  atUsers,
	}
}

// 点赞/取消点赞
type LikeReqAction uint8

const (
	LikeReqActionUndo LikeReqAction = 0
	LikeReqActionDo   LikeReqAction = 1
)
