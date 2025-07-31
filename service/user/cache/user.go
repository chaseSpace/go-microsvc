package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"microsvc/bizcomm/commuser"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/dao"
	"microsvc/service/user/deploy"
	"microsvc/util"
	"microsvc/util/db"
	"microsvc/util/utime"
	"time"

	"github.com/samber/lo"
)

type UserMap map[int64]*user.User

func GetUser(ctx context.Context, uid ...int64) (umap UserMap, err error) {
	var keys []string
	for _, u := range uid {
		keys = append(keys, fmt.Sprintf(UserInfoCacheKey, u))
	}
	var cacheMissUids []int64
	umap = make(UserMap, len(uid))

	if deploy.UserConf.DisableCache {
		cacheMissUids = uid
	} else {
		reply := user.R.MGet(ctx, keys...)
		if reply.Err() != nil {
			return nil, xerr.WrapRedis(reply.Err())
		}

		for i, v := range reply.Val() {
			if v == nil {
				cacheMissUids = append(cacheMissUids, uid[i])
				continue
			}
			u := new(user.User)
			_ = json.Unmarshal([]byte(v.(string)), u)
			if u.Id == 0 { // 可能缓存了UID=0的数据，过滤掉
				continue
			}
			umap[u.Uid] = u
		}
	}

	if len(cacheMissUids) > 0 {
		list, _, err := dao.GetUser(ctx, cacheMissUids...)
		if err != nil {
			return nil, err
		}
		lo.ForEach(list, func(item *user.User, index int) {
			umap[item.Uid] = item
			user.R.Set(ctx, fmt.Sprintf(UserInfoCacheKey, item.Uid), util.ToJson(item), UserInfoExpiry)
		})
		// 防止缓存穿透
		for _, id := range cacheMissUids {
			if umap[id] != nil {
				continue
			}
			umodel := commuser.NewUnknownUser(id)
			// TODO 在創建用戶時 刪除缓存
			user.R.Set(ctx, fmt.Sprintf(UserInfoCacheKey, id), util.ToJson(umodel), UserInfoExpiry)
		}
	}
	return
}

func GetOneUser(ctx context.Context, uid int64) (*user.User, error) {
	umap, err := GetUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	u := umap[uid]
	if u == nil {
		return nil, xerr.ErrUserNotFound.AppendMsg("uid=%d", uid)
	}
	return u, nil
}

func ClearUserInfo(ctx context.Context, uid int64) error {
	return user.R.Del(ctx, fmt.Sprintf(UserInfoCacheKey, uid)).Err()
}

type UserInfoUpdateController struct {
}

var UserInfoUpdateCtrl = UserInfoUpdateController{}

func (UserInfoUpdateController) DoesAllowUpdateUserInfo(ctx context.Context, body *userpb.UpdateBody, uid int64,
	rate *deploy.UpdateInfoRate) error {
	// 先判断开关
	if rate.Banned {
		return xerr.ErrForbidden.New("Feature not enabled")
	}
	key := fmt.Sprintf(UserInfoUpdateHistory, uid, body.FieldType.String())
	lastUpdate := user.R.LIndex(ctx, key, 0).Val()
	// 如果没有更新记录，则允许更新
	if lastUpdate == "" {
		return nil
	}
	trange, err := utime.CheckTimeStr(time.DateTime, lastUpdate)
	if err != nil {
		return err
	}
	lastUpdateAt := trange[0]
	//fmt.Println(33333, lastUpdateAt)
	// 检查时间期限限制
	if rate.DurationLimit != "" {
		t, err := time.ParseDuration(rate.DurationLimit)
		if err != nil {
			return err
		}
		if time.Now().Sub(lastUpdateAt) < t {
			//println(11111, fmt.Sprintf("%+v", new), time.Now().Sub(lastUpdateAt).String())
			return xerr.ErrForbidden.New("The operation is frequent")
		}
	}
	// 检查时间范围限制
	if len(rate.DateRangeLimit) == 2 {
		trange, err = utime.CheckTimeStr(time.DateTime, rate.DateRangeLimit...)
		if err != nil {
			return err
		}
		if !utime.IsInTimeRange(lastUpdateAt, trange[0], trange[1]) {
			//println(111112, fmt.Sprintf("%+v", new))
			return xerr.ErrForbidden.New("Feature not enabled. Please try again later")
		}
	}

	return nil
}

func (UserInfoUpdateController) AddUpdateInfoHistory(ctx context.Context, uid int64, infoType userpb.UserInfoType, maxHistoryLen int64) error {
	key := fmt.Sprintf(UserInfoUpdateHistory, uid, infoType.String())
	elem := time.Now().Format(time.DateTime)
	if maxHistoryLen < 1 {
		maxHistoryLen = 1
	}
	err := user.R.Eval(ctx, fmt.Sprintf(updateUserInfoHistoryLuaScript, maxHistoryLen), []string{key}, elem).Err()
	return db.IgnoreNilErr(err)
}
