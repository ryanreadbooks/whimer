package biz

import (
	"context"
	"encoding/json"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/misc/oss/signer"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type ImageAssetMetadata struct {
	Width  uint32 `json:"w"`
	Height uint32 `json:"h"`
	Format string `json:"f"`
}

type ImageAuth struct {
	ImageIds    []string
	CurrentTime int64
	ExpireTime  int64
	UploadAddr  string
	Token       string
}

// 评论资源管理
type AssetManagerBiz struct {
	inlineImageOssKeyGen *keygen.Generator           // 图片评论key生成
	imageUploadAuthGen   *signer.JwtUploadAuthSigner // 生成上传凭证
}

func NewAssetManagerBiz() *AssetManagerBiz {
	b := &AssetManagerBiz{
		inlineImageOssKeyGen: keygen.NewGenerator(
			keygen.WithBucket(config.Conf.Oss.InlineImage.Bucket),
			keygen.WithPrefix(config.Conf.Oss.InlineImage.Prefix),
			keygen.WithPrependBucket(true),
			keygen.WithPrependPrefix(true),
			keygen.WithStringer(keygen.RandomStringerV7{}),
		),
		imageUploadAuthGen: signer.NewJwtUploadAuthSigner(&config.Conf.OssUploadAuth),
	}

	return b
}

// 获取评论图片上传凭证
func (b *AssetManagerBiz) BatchGetImageAuths(ctx context.Context, cnt int32) (*ImageAuth, error) {
	keys := make([]string, 0, cnt)
	for range cnt {
		keys = append(keys, b.inlineImageOssKeyGen.Gen())
	}

	result, err := b.imageUploadAuthGen.BatchGetUploadAuth(keys, "comment_inline_image")
	if err != nil {
		return nil, xerror.Wrap(xerror.ErrServerSigning)
	}

	return &ImageAuth{
		ImageIds:    keys,
		CurrentTime: result.CurrentTime,
		ExpireTime:  result.ExpireTime,
		Token:       result.Token,
		UploadAddr:  config.Conf.Oss.InlineImage.UploadEndpoint,
	}, nil
}

func DoImageProxyUrl(asset *dao.CommentAsset) string {
	ak := config.Conf.ImgProxyAuth.GetKey()
	salt := config.Conf.ImgProxyAuth.GetSalt()
	if asset.Type == model.CommentAssetImage {
		// 加bucket
		return imgproxy.GetSignedUrl(
			config.Conf.Oss.InlineImage.DisplayEndpointBucket(), // 带bucket名
			asset.StoreKey,
			ak,
			salt,
			imgproxy.WithQuality("50"),
		)
	}

	return asset.StoreKey
}

func makeCommentAssetPO(commentId int64, images []*commentv1.CommentReqImage) []*dao.CommentAsset {
	assets := make([]*dao.CommentAsset, 0, len(images))
	for _, img := range images {
		meta := ImageAssetMetadata{
			Width:  img.Width,
			Height: img.Height,
			Format: img.Format,
		}
		metadata, _ := json.Marshal(&meta)
		assets = append(assets, &dao.CommentAsset{
			CommentId: commentId,
			Type:      model.CommentAssetImage,
			StoreKey:  img.StoreKey,
			Metadata:  metadata,
		})
	}

	return assets
}

func makePbCommentImage(assets []*dao.CommentAsset) []*commentv1.CommentItemImage {
	imgs := make([]*commentv1.CommentItemImage, 0, len(assets))
	for _, img := range assets {
		var meta ImageAssetMetadata
		_ = json.Unmarshal(img.Metadata, &meta)
		imgs = append(imgs, &commentv1.CommentItemImage{
			Url: DoImageProxyUrl(img),
			Meta: &commentv1.CommentItemImageMeta{
				Width:  meta.Width,
				Height: meta.Height,
				Format: meta.Format,
				Type:   model.CommentAssetType(img.Type).String(),
			},
		})
	}

	return imgs
}
