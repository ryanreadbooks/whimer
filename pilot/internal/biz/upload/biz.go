package upload

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	modelerr "github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"
)

type Biz struct {
	resourceDefine config.UploadResourceDefineMap
	uploaders      map[uploadresource.Type]*uploader
}

func NewBiz(c *config.Config) *Biz {
	b := &Biz{
		resourceDefine: c.UploadResourceDefineMap,
	}

	b.uploaders = make(map[uploadresource.Type]*uploader)
	for resourceType, metadata := range b.resourceDefine {
		b.uploaders[resourceType] = newUploader(&c.UploadAuthSign, &c.Oss, resourceType, metadata)
	}

	return b
}

func (b *Biz) getUploader(resource uploadresource.Type) (*uploader, error) {
	if uploader, ok := b.uploaders[resource]; ok {
		return uploader, nil
	}

	return nil, modelerr.ErrUnsupportedResource
}

func (b *Biz) RequestUploadAuth(ctx context.Context, resource uploadresource.Type, cnt int32, source string) (*StsTicket, error) {
	uploader, err := b.getUploader(resource)
	if err != nil {
		return nil, err
	}

	ticket, err := uploader.generateStsTicket(cnt, source)
	if err != nil {
		return nil, xerror.Wrapf(err, "uploader generate ticket failed").WithCtx(ctx)
	}

	return ticket, nil
}
