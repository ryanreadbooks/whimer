package assetprocess

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type VideoProcessor struct {
	biz biz.Biz
}

func newVideoProcessor(biz biz.Biz) Processor {
	return &VideoProcessor{biz: biz}
}

func (p *VideoProcessor) Process(ctx context.Context, note *model.Note) error {
	return nil
}
