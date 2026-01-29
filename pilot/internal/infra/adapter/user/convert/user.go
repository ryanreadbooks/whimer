package convert

import (
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	uservo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
)

func PbUserInfoToVoUser(pb *userv1.UserInfo) *uservo.User {
	return &uservo.User{
		Uid:       pb.GetUid(),
		Nickname:  pb.GetNickname(),
		Avatar:    pb.GetAvatar(),
		StyleSign: pb.GetStyleSign(),
		Gender:    pb.GetGender(),
	}
}
