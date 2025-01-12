package profile

import "github.com/ryanreadbooks/whimer/misc/xerror"

type HoverReq struct {
	UserId uint64 `form:"user_id"`
}

func (r *HoverReq) Validate() error {
	if r.UserId == 0 {
		return xerror.ErrArgs.Msg("用户不存在")
	}

	return nil
}

type HoverRes struct {
	BasicInfo struct {
		Nickname  string `json:"nickname"`
		StyleSign string `json:"style_sign"`
		Avatar    string `json:"avatar"`
	} `json:"basic_info"`

	Interaction struct {
		Fans    string `json:"fans"`
		Follows string `json:"follows"`
	} `json:"interaction"`

	Relation struct {
		Status string `json:"status"`
	} `json:"relation"`
}

const (
	RelationFollowing = "following"
	RelationNone      = "none"
)

// 模糊化处理数字
func HideActualCount(cnt uint64) string {
	if cnt == 0 {
		return "0"
	}

	if cnt >= 1 && cnt < 10 {
		return "1+"
	}

	if cnt >= 10 && cnt < 100 {
		return "10+"
	}

	if cnt >= 100 && cnt < 1000 {
		return "100+"
	}

	if cnt >= 1000 && cnt < 10000 {
		return "1K+"
	}

	if cnt >= 10000 && cnt < 100000 {
		return "1W+"
	}

	if cnt >= 100000 && cnt < 1000000 {
		return "10W+"
	}

	return "100W+"
}
