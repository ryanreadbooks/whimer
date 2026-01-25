package user

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/ryanreadbooks/whimer/pilot/internal/app/user/dto"
	noteentity "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	noterepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	relationrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/repository"
	userdomain "github.com/ryanreadbooks/whimer/pilot/internal/domain/user"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"

	"golang.org/x/sync/errgroup"
)

type Service struct {
	userDomainService  *userdomain.DomainService
	userAdapter        repository.UserServiceAdapter
	relationAdapter    relationrepo.RelationAdapter
	noteCreatorAdapter noterepo.NoteCreatorAdapter
	noteFeedAdapter    noterepo.NoteFeedAdapter
	userSettingRepo    repository.UserSettingRepository
	recentContactRepo  repository.RecentContactRepository
}

func NewService(
	userDomainService *userdomain.DomainService,
	userAdapter repository.UserServiceAdapter,
	relationAdapter relationrepo.RelationAdapter,
	noteCreatorAdapter noterepo.NoteCreatorAdapter,
	noteFeedAdapter noterepo.NoteFeedAdapter,
	userSettingRepo repository.UserSettingRepository,
	recentContactRepo repository.RecentContactRepository,
) *Service {
	return &Service{
		userDomainService:  userDomainService,
		userAdapter:        userAdapter,
		relationAdapter:    relationAdapter,
		noteCreatorAdapter: noteCreatorAdapter,
		noteFeedAdapter:    noteFeedAdapter,
		userSettingRepo:    userSettingRepo,
		recentContactRepo:  recentContactRepo,
	}
}

func (s *Service) ListUsers(ctx context.Context, uids []int64) (map[int64]*dto.User, error) {
	users, err := s.userAdapter.BatchGetUser(ctx, uids)
	if err != nil {
		return nil, xerror.Wrapf(err, "user adapter batch get user failed").WithCtx(ctx)
	}

	result := make(map[int64]*dto.User, len(users))
	for uid, user := range users {
		result[uid] = dto.ConvertVoUserToDto(user)
	}

	return result, nil
}

func (s *Service) GetUser(ctx context.Context, uid int64) (*dto.User, error) {
	user, err := s.userAdapter.GetUser(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "user adapter get user failed").WithCtx(ctx)
	}

	return dto.ConvertVoUserToDto(user), nil
}

// 获取用户的投稿数量、点赞数量等信息
func (s *Service) GetUserStat(ctx context.Context, targetUid int64) (*dto.UserStat, error) {
	reqUid := metadata.Uid(ctx)
	stat := vo.UserStat{}

	eg, ctx := errgroup.WithContext(ctx)

	// 用户投稿数量
	eg.Go(func() error {
		return recovery.Do(func() error {
			var cnt int64
			var err error

			if reqUid == targetUid {
				cnt, err = s.noteCreatorAdapter.GetPostedCount(ctx, reqUid)
			} else {
				cnt, err = s.noteFeedAdapter.GetPublicPostedCount(ctx, targetUid)
			}
			if err != nil {
				return err
			}
			stat.Posted = cnt
			return nil
		})
	})

	// 用户粉丝数量
	eg.Go(func() error {
		return recovery.Do(func() error {
			cnt, err := s.relationAdapter.GetFanCount(ctx, targetUid)
			if err != nil {
				return err
			}
			stat.Fans = cnt
			return nil
		})
	})

	// 用户关注数量
	eg.Go(func() error {
		return recovery.Do(func() error {
			cnt, err := s.relationAdapter.GetFollowingCount(ctx, targetUid)
			if err != nil {
				return err
			}
			stat.Followings = cnt
			return nil
		})
	})

	if err := eg.Wait(); err != nil {
		return dto.ConvertVoUserStatToDto(&stat), err
	}

	return dto.ConvertVoUserStatToDto(&stat), nil
}

// 获取用户卡片信息
func (s *Service) GetHoverProfile(ctx context.Context, targetUid int64) (*dto.HoverInfo, error) {
	uid := metadata.Uid(ctx)
	isAuthedRequest := uid != 0

	eg, ctx := errgroup.WithContext(ctx)
	var targetUser *vo.User

	// 基本信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			res, err := s.userAdapter.GetUser(ctx, targetUid)
			if err != nil {
				return err
			}
			targetUser = res
			return nil
		})
	})

	var fansCount, followsCount int64
	// 用户的粉丝，关注等信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			fansCount, err = s.relationAdapter.GetFanCount(ctx, targetUid)
			if err != nil {
				return err
			}
			followsCount, err = s.relationAdapter.GetFollowingCount(ctx, targetUid)
			return err
		})
	})

	// 用户最近发布的笔记信息
	var postAssets []*noteentity.RecentPost
	eg.Go(func() error {
		return recovery.Do(func() error {
			posts, err := s.noteFeedAdapter.GetUserRecentPost(ctx, targetUid, 3)
			if err != nil {
				return err
			}
			postAssets = posts
			return nil
		})
	})

	var followed bool
	if isAuthedRequest {
		eg.Go(func() error {
			return recovery.Do(func() error {
				var err error
				followed, err = s.relationAdapter.CheckFollowed(ctx, uid, targetUid)
				return err
			})
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// 组织结果
	hoverInfo := &vo.HoverInfo{
		Relation: vo.HoverRelation{Status: vo.RelationNone},
	}
	hoverInfo.BasicInfo.Nickname = targetUser.Nickname
	hoverInfo.BasicInfo.StyleSign = targetUser.StyleSign
	hoverInfo.BasicInfo.Avatar = targetUser.Avatar
	hoverInfo.Interaction.Fans = strconv.FormatInt(fansCount, 10)
	hoverInfo.Interaction.Follows = strconv.FormatInt(followsCount, 10)

	if followed {
		hoverInfo.Relation.Status = vo.RelationFollowing
	}
	hoverInfo.RecentPosts = postAssets

	return dto.ConvertVoHoverInfoToDto(hoverInfo, isAuthedRequest), nil
}

type uidAndTime struct {
	Uid  int64
	Time int64
}

type followingUser struct {
	*vo.User
	followTime int64
}

// 按照nickname获取关注的用户
//
// 由于是不同服务存储关注关系和用户信息 所以此种方法可能在数据量大的时候很慢
//
// 这里只提供获取关注的方法不提供获取粉丝的方法 因为如果按照这种方式获取粉丝数量的话 在粉丝量庞大时会导致性能问题;
// 由于限制了关注的人数 所以暴力获取的方式应该能接受
func (s *Service) BrutalListFollowingsByName(ctx context.Context, uid int64, target string) ([]*vo.User, error) {
	var (
		offset int64 = 0
		count  int32 = 250
	)

	followings := make([]uidAndTime, 0, 128)
	for {
		result, err := s.relationAdapter.GetUserFollowingList(ctx, uid, offset, count)
		if err != nil {
			return nil, xerror.Wrapf(err, "get user following list failed")
		}

		if len(result.Followings) == 0 {
			break
		}

		for i := range result.Followings {
			followings = append(followings, uidAndTime{
				Uid:  result.Followings[i],
				Time: result.FollowTimes[i],
			})
		}

		if result.HasMore {
			offset = result.NextOffset
		} else {
			break
		}
	}

	if len(followings) == 0 {
		return []*vo.User{}, nil
	}

	followingsMap := xslice.MakeMap(followings, func(v uidAndTime) int64 {
		return v.Uid
	})

	uids := xmap.Keys(followingsMap)
	users, err := s.userAdapter.BatchGetUser(ctx, uids)
	if err != nil {
		return nil, xerror.Wrapf(err, "batch get user failed")
	}

	// 本地筛选nickname
	resultUsers := make([]*vo.User, 0, len(users))
	for _, user := range users {
		if strings.Contains(user.Nickname, target) {
			resultUsers = append(resultUsers, user)
		}
	}

	resultUsersMap := xslice.MakeMap(resultUsers, func(v *vo.User) int64 {
		return v.Uid
	})

	followingUsers := make([]*followingUser, 0, len(resultUsersMap))
	for _, user := range resultUsersMap {
		followingUsers = append(followingUsers, &followingUser{
			User:       user,
			followTime: followingsMap[user.Uid].Time,
		})
	}

	// 按照关注时间排序
	sort.Slice(followingUsers, func(i, j int) bool {
		return followingUsers[i].followTime > followingUsers[j].followTime
	})

	results := make([]*vo.User, 0, len(followingUsers))
	for _, user := range followingUsers {
		results = append(results, user.User)
	}

	return results, nil
}

// 获取@用户候选列表
func (s *Service) GetMentionUserCandidates(ctx context.Context, uid int64, search string) ([]*dto.MentionUserItem, error) {
	eg, ctx := errgroup.WithContext(ctx)

	groups := make([]*vo.MentionUserItem, 3)

	// 拿最近联系人
	eg.Go(recovery.DoV2(func() error {
		recentContacts, err := s.userDomainService.GetAllRecentContacts(ctx, uid)
		if err != nil {
			xlog.Msg("get recent contacts failed").Err(err).Errorx(ctx)
		}

		groups[0] = &vo.MentionUserItem{
			Group:     vo.MentionRecentContacts,
			GroupDesc: vo.MentionRecentContacts.Desc(),
			Users:     recentContacts,
		}
		return nil
	}))

	// 我的关注
	eg.Go(recovery.DoV2(func() error {
		myFollowings, err := s.BrutalListFollowingsByName(ctx, uid, search)
		if err != nil {
			xlog.Msg("list followings groups failed").Err(err).Errorx(ctx)
		}

		groups[1] = &vo.MentionUserItem{
			Group:     vo.MentionFollowings,
			GroupDesc: vo.MentionFollowings.Desc(),
			Users:     myFollowings,
		}
		return nil
	}))

	// TODO 其他人
	if len(search) > 0 {
		eg.Go(recovery.DoV2(func() error {
			groups[2] = &vo.MentionUserItem{
				Group:     vo.MentionOthers,
				GroupDesc: vo.MentionOthers.Desc(),
				Users:     []*vo.User{},
			}
			return nil
		}))
	}

	if err := eg.Wait(); err != nil {
		return nil, xerror.Wrapf(err, "get mention user candidates failed").WithCtx(ctx)
	}

	result := make([]*dto.MentionUserItem, 0, len(groups))
	for _, g := range groups {
		if g != nil {
			result = append(result, dto.ConvertVoMentionItemToDto(g))
		}
	}

	return result, nil
}
