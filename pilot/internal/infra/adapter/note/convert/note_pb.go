package convert

import (
	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

func EntityCreateNoteParamsAsPb(params *entity.CreateNoteParams) *notev1.CreateNoteRequest {
	imageReqs := make([]*notev1.CreateReqImage, 0, len(params.Images))
	for _, image := range params.Images {
		imageReqs = append(imageReqs, &notev1.CreateReqImage{
			FileId: image.FileId,
			Width:  image.Width,
			Height: image.Height,
			Format: image.Format,
		})
	}

	var videoReq *notev1.CreateReqVideo
	if params.Videos != nil {
		videoReq = &notev1.CreateReqVideo{
			FileId:       params.Videos.FileId,
			CoverFileId:  params.Videos.CoverFileId,
			TargetFileId: params.Videos.GetTargetFileId(),
		}
	}

	// atUsers
	atUsers := make([]*notev1.NoteAtUser, 0, len(params.AtUsers))
	for _, user := range params.AtUsers {
		atUsers = append(atUsers, &notev1.NoteAtUser{
			Uid:      user.Uid,
			Nickname: user.Nickname,
		})
	}

	// tagList
	tagList := make([]int64, 0, len(params.Tags))
	for _, tag := range params.Tags {
		tagList = append(tagList, tag.Id)
	}

	req := &notev1.CreateNoteRequest{
		Basic: &notev1.CreateReqBasic{
			Title:     params.Title,
			Desc:      params.Desc,
			Privacy:   NotePrivacyAsPb(params.Privacy),
			AssetType: NoteAssetTypeAsPb(params.AssetType),
		},
		Images:  imageReqs,
		Video:   videoReq,
		AtUsers: atUsers,
		Tags: &notev1.CreateReqTag{
			TagList: tagList,
		},
	}

	return req
}

func PbNoteImageToEntity(image *notev1.NoteImage) *entity.NoteImage {
	if image == nil {
		return nil
	}

	return &entity.NoteImage{
		FileId: image.GetKey(),
		Width:  image.GetMeta().GetWidth(),
		Height: image.GetMeta().GetHeight(),
		Format: image.GetMeta().GetFormat(),
	}
}

func BatchPbNoteImagesToEntities(images []*notev1.NoteImage) []*entity.NoteImage {
	if len(images) == 0 {
		return nil
	}

	entities := make([]*entity.NoteImage, 0, len(images))
	for _, image := range images {
		entities = append(entities, PbNoteImageToEntity(image))
	}
	return entities
}

func PbNoteVideoToEntity(video *notev1.NoteVideo) *entity.NoteVideo {
	if video == nil {
		return nil
	}

	v := &entity.NoteVideo{
		FileId: video.GetKey(),
	}

	v.SetMetadata(&entity.NoteVideoMetadata{
		Width:      video.GetMeta().GetWidth(),
		Height:     video.GetMeta().GetHeight(),
		Format:     video.GetMeta().GetFormat(),
		Duration:   video.GetMeta().GetDuration(),
		Bitrate:    video.GetMeta().GetBitrate(),
		Codec:      video.GetMeta().GetCodec(),
		Framerate:  video.GetMeta().GetFramerate(),
		AudioCodec: video.GetMeta().GetAudioCodec(),
	})

	return v
}

func BatchPbNoteVideosToEntities(videos []*notev1.NoteVideo) []*entity.NoteVideo {
	if len(videos) == 0 {
		return nil
	}

	entities := make([]*entity.NoteVideo, 0, len(videos))
	for _, video := range videos {
		entities = append(entities, PbNoteVideoToEntity(video))
	}

	return entities
}

func PbNoteToCreatorEntity(note *notev1.NoteItem) *entity.CreatorNote {
	if note == nil {
		return nil
	}

	return &entity.CreatorNote{
		Id:         vo.NoteId(note.GetNoteId()),
		Title:      note.GetTitle(),
		Desc:       note.GetDesc(),
		Privacy:    PbNotePrivacyToVo(note.GetPrivacy()),
		AssetType:  PbNoteAssetTypeToVo(note.GetNoteType()),
		Status:     PbLifeCycleStateToNoteStatus(note.GetLifeCycleState()),
		Type:       PbNoteAssetTypeToVo(note.NoteType).AsNoteType(),
		Ip:         note.Ip,
		CreateTime: note.GetCreateAt(),
		UpdateTime: note.GetUpdateAt(),
		Images:     BatchPbNoteImagesToEntities(note.GetImages()),
		Videos:     BatchPbNoteVideosToEntities(note.GetVideos()),
		Tags:       BatchPbNoteTagsToEntities(note.GetTags()),
		AtUsers:    BatchPbAtUsersToVos(note.GetAtUsers()),
		Likes:      note.Likes,
		Replies:    note.Replies,
		OwnerId:    note.Owner,
	}
}

// BatchPbNoteTagsToEntities 转换 pb NoteTags 到 entity NoteTags
func BatchPbNoteTagsToEntities(tags []*notev1.NoteTag) []*entity.NoteTag {
	if len(tags) == 0 {
		return nil
	}
	entities := make([]*entity.NoteTag, 0, len(tags))
	for _, tag := range tags {
		entities = append(entities, &entity.NoteTag{
			Id:   tag.GetId(),
			Name: tag.GetName(),
		})
	}
	return entities
}

// BatchPbAtUsersToVos 转换 pb NoteAtUsers 到 vo AtUserList
func BatchPbAtUsersToVos(atUsers []*notev1.NoteAtUser) mentionvo.AtUserList {
	if len(atUsers) == 0 {
		return nil
	}
	vos := make(mentionvo.AtUserList, 0, len(atUsers))
	for _, u := range atUsers {
		vos = append(vos, &mentionvo.AtUser{
			Uid:      u.GetUid(),
			Nickname: u.GetNickname(),
		})
	}
	return vos
}

// PbLifeCycleStateToNoteStatus 转换 pb NoteLifeCycleState 到 vo NoteStatus
func PbLifeCycleStateToNoteStatus(state notev1.NoteLifeCycleState) vo.NoteStatus {
	switch state {
	case notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_PUBLISHED:
		return vo.NoteStatusPublished
	case notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_AUDITING:
		return vo.NoteStatusAuditing
	case notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_BANNED:
		return vo.NoteStatusBanned
	case notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_REJECTED:
		return vo.NoteStatusRejected
	default:
		return vo.NoteStatusUnknown
	}
}

func PbNotePrivacyToVo(privacy int32) vo.Visibility {
	switch privacy {
	case int32(notev1.NotePrivacy_PUBLIC):
		return vo.VisibilityPublic
	case int32(notev1.NotePrivacy_PRIVATE):
		return vo.VisibilityPrivate
	default:
		return vo.Visibility(notev1.NotePrivacy_NOTE_PRIVACY_UNSPECIFIED)
	}
}

func NotePrivacyAsPb(privacy vo.Visibility) int32 {
	switch privacy {
	case vo.VisibilityPublic:
		return int32(notev1.NotePrivacy_PUBLIC)
	case vo.VisibilityPrivate:
		return int32(notev1.NotePrivacy_PRIVATE)
	}
	return int32(notev1.NotePrivacy_NOTE_PRIVACY_UNSPECIFIED)
}

func PbNoteAssetTypeToVo(assetType notev1.NoteAssetType) vo.AssetType {
	switch assetType {
	case notev1.NoteAssetType_IMAGE:
		return vo.AssetTypeImage
	case notev1.NoteAssetType_VIDEO:
		return vo.AssetTypeVideo
	default:
		return vo.AssetType(notev1.NoteAssetType_NOTE_ASSET_TYPE_UNSPECIFIED)
	}
}

func NoteAssetTypeAsPb(assetType vo.AssetType) notev1.NoteAssetType {
	switch assetType {
	case vo.AssetTypeImage:
		return notev1.NoteAssetType_IMAGE
	case vo.AssetTypeVideo:
		return notev1.NoteAssetType_VIDEO
	}
	return notev1.NoteAssetType(notev1.NoteAssetType_NOTE_ASSET_TYPE_UNSPECIFIED)
}

// NoteStatusToLifeCycleState 将 domain NoteStatus 转换为 pb NoteLifeCycleState
func NoteStatusToLifeCycleState(status vo.NoteStatus) notev1.NoteLifeCycleState {
	switch status {
	case vo.NoteStatusPublished:
		return notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_PUBLISHED
	case vo.NoteStatusAuditing:
		return notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_AUDITING
	case vo.NoteStatusBanned:
		return notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_BANNED
	case vo.NoteStatusRejected:
		return notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_REJECTED
	default:
		return notev1.NoteLifeCycleState_NOTE_LIFE_CYCLE_STATE_UNSPECIFIED
	}
}

// BatchPbNotesToCreatorEntities 批量转换 pb NoteItem 到 entity CreatorNote
func BatchPbNotesToCreatorEntities(items []*notev1.NoteItem) []*entity.CreatorNote {
	if len(items) == 0 {
		return []*entity.CreatorNote{}
	}
	entities := make([]*entity.CreatorNote, 0, len(items))
	for _, item := range items {
		entities = append(entities, PbNoteToCreatorEntity(item))
	}
	return entities
}

func LikeActionAsPbLikeOperation(action vo.LikeAction) notev1.LikeNoteRequest_Operation {
	switch action {
	case vo.LikeActionDo:
		return notev1.LikeNoteRequest_OPERATION_DO_LIKE
	default:
		return notev1.LikeNoteRequest_OPERATION_UNDO_LIKE
	}
}

func PbFeedNoteToEntity(note *notev1.FeedNoteItem) *entity.FeedNote {
	if note == nil {
		return nil
	}

	return &entity.FeedNote{
		Id:        vo.NoteId(note.GetNoteId()),
		Title:     note.GetTitle(),
		Desc:      note.GetDesc(),
		CreateAt:  note.GetCreatedAt(),
		UpdateAt:  note.GetUpdatedAt(),
		Images:    BatchPbNoteImagesToEntities(note.GetImages()),
		Videos:    BatchPbNoteVideosToEntities(note.GetVideos()),
		Likes:     note.GetLikes(),
		Comments:  note.GetReplies(),
		Ip:        note.GetIp(),
		Type:      PbNoteAssetTypeToVo(note.GetNoteType()).AsNoteType(),
		AuthorUid: note.GetAuthor(),
	}
}

func PbFeedNoteExtToEntity(ext *notev1.FeedNoteItemExt) *entity.FeedNoteExt {
	if ext == nil {
		return nil
	}

	return &entity.FeedNoteExt{
		Tags:    BatchPbNoteTagsToEntities(ext.GetTags()),
		AtUsers: BatchPbAtUsersToVos(ext.GetAtUsers()),
	}
}

func BatchPbFeedNotesToEntities(notes []*notev1.FeedNoteItem) []*entity.FeedNote {
	if len(notes) == 0 {
		return nil
	}
	entities := make([]*entity.FeedNote, 0, len(notes))
	for _, note := range notes {
		entities = append(entities, PbFeedNoteToEntity(note))
	}
	return entities
}

func NewNoteImageItemUrlPrv(pbimg *notev1.NoteImage) string {
	noteAssetBucket := config.Conf.UploadResourceDefineMap["note_image"].Bucket

	url := imgproxy.GetSignedUrl(
		config.Conf.Oss.DisplayEndpointBucket(noteAssetBucket),
		pbimg.GetKey(),
		config.Conf.ImgProxyAuth.GetKey(),
		config.Conf.ImgProxyAuth.GetSalt(),
		imgproxy.WithQuality(config.Conf.ImgQuality.QualityPreview))
	return url
}

func PbNoteTypeToVoNoteType(n notev1.NoteAssetType) vo.NoteType {
	switch n {
	case notev1.NoteAssetType_IMAGE:
		return vo.NoteTypeImage
	case notev1.NoteAssetType_VIDEO:
		return vo.NoteTypeVideo
	default:
		return ""
	}
}
