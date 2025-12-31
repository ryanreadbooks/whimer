package user

import usermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user/model"

type MentionUserReq struct {
	Search string `form:"search,optional"`
}

type MentionUserResp struct {
	Groups []*usermodel.MentionUserRespItem `json:"groups"`
}

type SetNoteShowSettingsReq struct {
	ShowNoteLikes bool `json:"show_note_likes"`
}
