package biz

import (
	"time"

	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

func formatNoteVideoAsset(req *CreateNoteRequestVideo) []*notedao.AssetPO {
	assets := make([]*notedao.AssetPO, 3)
	now := time.Now().Unix()
	// 此时暂时无法得知meta信息 需要后续处理完再回填进数据库中
	// h264
	assets[0] = &notedao.AssetPO{
		AssetKey:  model.FormatNoteVideoKey(req.TargetFileId, model.SupportedVideoH264Suffix),
		AssetType: model.AssetTypeVideo,
		CreateAt:  now,
	}

	// // h265
	// assets[1] = &notedao.AssetPO{
	// 	AssetKey:  model.FormatNoteVideoKey(req.TargetFileId, model.SupportedVideoH265Suffix),
	// 	AssetType: model.AssetTypeVideo,
	// 	CreateAt:  now,
	// }

	// av1
	assets[1] = &notedao.AssetPO{
		AssetKey:  model.FormatNoteVideoKey(req.TargetFileId, model.SupportedVideoAV1Suffix),
		AssetType: model.AssetTypeVideo,
		CreateAt:  now,
	}

	return assets
}
