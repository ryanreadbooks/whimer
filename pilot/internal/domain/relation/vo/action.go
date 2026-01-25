package vo

type FollowAction int8

const (
	ActionFollow   FollowAction = 1
	ActionUnFollow FollowAction = 2
)

func (a FollowAction) IsValid() bool {
	return a == ActionFollow || a == ActionUnFollow
}
