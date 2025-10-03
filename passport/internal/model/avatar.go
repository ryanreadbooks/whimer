package model

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/passport/internal/global"

	"github.com/gabriel-vasile/mimetype"
)

// 校验上传的文件元信息
type AvatarInfoRequest struct {
	Filename    string `json:"filename"`
	Ext         string `json:"ext"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Content     []byte `json:"-"`
}

func (r *AvatarInfoRequest) String() string {
	s, _ := json.Marshal(r)
	return utils.Bytes2String(s)
}

func ParseAvatarFile(file multipart.File, header *multipart.FileHeader) (*AvatarInfoRequest, error) {
	if header == nil {
		return nil, global.ErrArgs
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, global.ErrUploadAvatar
	}

	if len(content) > MaxAvatarBytes {
		return nil, global.ErrAvatarTooLarge
	}

	// 检测mimeType
	detectedMime := mimetype.Detect(content)
	if !avatarFormatSupported(detectedMime.String()) {
		return nil, global.ErrAvatarFormatUnsupported.ExtMsg(detectedMime.String())
	}

	info := &AvatarInfoRequest{
		Filename:    fmt.Sprintf("%d%s", time.Now().Unix(), detectedMime.Extension()), // 生成临时文件名
		Ext:         detectedMime.Extension(),
		ContentType: detectedMime.String(),
		Size:        int64(len(content)),
		Content:     content,
	}

	return info, nil
}

// 定义avatar支持上传的格式
func avatarFormatSupported(mimeType string) bool {
	return mimeType == "image/jpeg" ||
		mimeType == "image/png" ||
		mimeType == "image/webp"
}

// 上传头像响应结果
type UploadUserAvatarResponse struct {
	Url string `json:"avatar_url"`
}
