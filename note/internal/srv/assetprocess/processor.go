package assetprocess

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type Processor interface {
	Process(ctx context.Context, note *model.Note) error
}


func NewProcessor(noteType model.NoteType, biz biz.Biz) Processor {
	switch noteType {
	case model.AssetTypeImage:
		return newImageProcessor(biz)
	case model.AssetTypeVideo:
		return newVideoProcessor(biz)
	}
	return nil
}
