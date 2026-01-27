package upload

import (
	"slices"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"
)

type GetTempCredsReq struct {
	Resource string `form:"resource"`
	Source   string `form:"source,optional"`
	Count    int32  `form:"count,optional"`
}

var (
	allowedTempCredsResourceType = map[string]struct{}{
		string(uploadresource.NoteImage): {},
		string(uploadresource.NoteVideo): {},
	}

	allowedPostPolicyResourceType = map[string]struct{}{
		string(uploadresource.NoteVideoCover): {},
	}
)

const (
	maxCountOfVideoUploads = 1
	maxCountOfImageUploads = 8
)

func (r *GetTempCredsReq) Validate() error {
	if r.Count <= 0 {
		return xerror.ErrInvalidArgs.Msg("参数错误")
	}

	if _, ok := allowedTempCredsResourceType[r.Resource]; !ok {
		return xerror.ErrInvalidArgs.Msg("不支持的资源类型")
	}

	if r.Resource == string(uploadresource.NoteImage) {
		if r.Count > maxCountOfImageUploads {
			return xerror.ErrInvalidArgs.Msg("不支持请求这么多上传凭证")
		}
	} else if r.Resource == string(uploadresource.NoteVideo) {
		if r.Count > maxCountOfVideoUploads {
			return xerror.ErrInvalidArgs.Msg("不支持请求这么多上传凭证")
		}
	}

	return nil
}

type GetTempCredsResp struct {
	UploadFile  UploadFile  `json:"upload_file"`
	UploadCreds UploadCreds `json:"upload_creds"`
}

type UploadFile struct {
	Bucket string   `json:"bucket"`
	Ids    []string `json:"ids"`
}

type UploadCreds struct {
	TmpAccessKey string `json:"tmp_access_key"`
	TmpSecretKey string `json:"tmp_secret_key"`
	SessionToken string `json:"session_token"`
	ExpireAt     int64  `json:"expire_at"` // unix timestamp in second
	UploadAddr   string `json:"upload_addr"`
}

type GetPostPolicyCredsReq struct {
	Resource string `form:"resource"`
	MimeType string `form:"mime_type"` // mime type
	Sha256   string `form:"sha256"`
	Size     int64  `form:"size"` // in bytes
}

func (r *GetPostPolicyCredsReq) Validate() error {
	if _, ok := allowedPostPolicyResourceType[r.Resource]; !ok {
		return xerror.ErrInvalidArgs.Msg("不支持的资源类型")
	}

	if r.Size <= 0 {
		return xerror.ErrInvalidArgs.Msg("非法大小")
	}

	if r.Sha256 == "" {
		return xerror.ErrInvalidArgs.Msg("非法sha256")
	}

	if !slices.Contains(uploadresource.NoteVideoCover.PermitContentType(), r.MimeType) {
		return xerror.ErrInvalidArgs.Msg("不支持的封面类型")
	}

	if r.Size > uploadresource.NoteVideoCover.PermitSize() {
		return xerror.ErrInvalidArgs.Msg("封面大小超过限制")
	}

	return nil
}

type GetPostPolicyCredsResp struct {
	FileId     string            `json:"file_id"`
	UploadAddr string            `json:"upload_addr"`
	Form       map[string]string `json:"form"`
}
