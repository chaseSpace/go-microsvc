package cache

import (
	"context"
	"fmt"
	"microsvc/model/svc/friend"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/friendpb"
	"microsvc/service/friend/dao"
	"microsvc/util/db"
	"microsvc/util/utime"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type friendCtrlT struct {
	friendListExpire time.Duration
	onewayListExpire time.Duration
}

var FriendCtrl = friendCtrlT{
	friendListExpire: time.Minute * 5,
	onewayListExpire: time.Minute * 5,
}

func (friendCtrlT) friendListKey(uid int64) string {
	return fmt.Sprintf(friendListCacheKey, uid)
}

func (friendCtrlT) onewayListKey(uid int64, isFollow bool) string {
	if isFollow {
		return fmt.Sprintf(followListCacheKey, uid)
	}
	return fmt.Sprintf(fansListCacheKey, uid)
}

func (f friendCtrlT) FriendList(ctx context.Context, uid int64, req *friendpb.FriendListReq) (list []*friend.FriendRPC, total int64, err error) {
	var buf []byte
	var listDao []*friend.Friend
	var ckey = f.friendListKey(uid)

	defer func() {
		// 用户信息每次都要重新获取（有单独的用户信息缓存）
		list, err = SetupFriendUserInfo(ctx, listDao)
	}()

	// 仅缓存第一页（缓存全部容易有bug，且意义不大）
	if req.Page.Pn == 1 {
		buf, err = friend.R.Get(ctx, ckey).Bytes()
		if db.IgnoreNilErr(err) != nil {
			return
		}
		if buf != nil {
			var model friendListT
			_ = jsoniter.Unmarshal(buf, &model)
			if model.Total == 0 {
				return
			}
			total = model.Total
			listDao = model.List
			return
		}
	}

	// 非第一页 或 缓存不存在，从数据库获取
	listDao, total, err = dao.FriendList(ctx, uid, req)
	if err != nil {
		return
	}
	if req.Page.Pn > 1 {
		return
	}
	// 缓存第一页
	if len(listDao) == 0 {
		buf = []byte("{}")
	} else {
		buf, _ = jsoniter.Marshal(&friendListT{List: listDao, Total: total})
	}
	err = friend.R.Set(ctx, ckey, buf, f.friendListExpire).Err()
	return
}

func (f friendCtrlT) OnewayList(ctx context.Context, uid int64, req *friendpb.FriendOnewayListReq) (list []*friend.FriendRPC, total int64, err error) {
	var buf []byte
	var listDao []*friend.Friend
	var ckey = f.onewayListKey(uid, req.IsFollow)

	defer func() {
		// 用户信息每次都要重新获取（有单独的用户信息缓存）
		list, err = SetupFriendUserInfo(ctx, listDao)
	}()

	// 仅缓存第一页（缓存全部容易有bug，且意义不大）
	if req.Page.Pn == 1 {
		buf, err = friend.R.Get(ctx, ckey).Bytes()
		if db.IgnoreNilErr(err) != nil {
			return
		}
		if buf != nil {
			var model friendListT
			_ = jsoniter.Unmarshal(buf, &model)
			if model.Total == 0 {
				return
			}
			total = model.Total
			listDao = model.List
			return
		}
	}

	// 非第一页 或 缓存不存在，从数据库获取
	listDao, total, err = dao.OnewayList(ctx, uid, req)
	if err != nil {
		return
	}

	if req.Page.Pn > 1 {
		return
	}

	// 缓存第一页
	// 空也缓存，避免缓存穿透
	if len(listDao) == 0 {
		buf = []byte("{}")
	} else {
		buf, _ = jsoniter.Marshal(&friendListT{List: listDao, Total: total})
	}
	err = friend.R.Set(ctx, ckey, buf, f.onewayListExpire).Err()
	return
}

func (f friendCtrlT) CleanFriendList(ctx context.Context, uids ...int64) error {
	for _, uid := range uids {
		err := friend.R.Del(ctx, f.friendListKey(uid)).Err()
		if err != nil {
			return xerr.WrapRedis(err)
		}
	}
	return nil
}

func (f friendCtrlT) CleanOnewayList(ctx context.Context, uid, peerUid int64, isFollow bool) error {
	key := f.onewayListKey(uid, isFollow)
	err := friend.R.Del(ctx, key).Err()
	if err != nil {
		return xerr.WrapRedis(err)
	}
	if isFollow {
		// 我的关注 + 对方的粉丝列表缓存
		key = f.onewayListKey(peerUid, !isFollow)
	} else {
		// 我的粉丝 + 对方的关注列表缓存
		key = f.onewayListKey(peerUid, isFollow)
	}
	err = friend.R.Del(ctx, key).Err()
	return xerr.WrapRedis(err)
}

type visitorCtrlT struct {
	visitorSaveTsExpire time.Duration
	visitorListExpire   time.Duration
}

var VisitorCtrl = visitorCtrlT{
	visitorSaveTsExpire: time.Minute * 3,
	visitorListExpire:   time.Minute * 5,
}

func (v *visitorCtrlT) visitorSaveKey(uid, targetId int64) string {
	return fmt.Sprintf(saveVisitorCacheKey, uid, targetId)
}
func (v *visitorCtrlT) visitorListKey(uid int64) string {
	return fmt.Sprintf(visitorListCacheKey, uid)
}

func (v *visitorCtrlT) AllowSaveVisitor(ctx context.Context, uid, targetId int64, seconds *int64) (allow bool, err error) {
	if *seconds < 1 {
		return false, xerr.ErrParams.New("时间不能小于1秒")
	}
	key := v.visitorSaveKey(uid, targetId)

	// 过大的时间会被截取
	if expSeconds := int64(v.visitorSaveTsExpire.Seconds()); *seconds > expSeconds {
		*seconds = expSeconds
	}
	var lastVisitorSaveTs int64
	lastVisitorSaveTs, err = friend.R.Get(ctx, key).Int64()
	if db.IgnoreNilErr(err) != nil {
		return
	}

	now := time.Now().Unix()
	if lastVisitorSaveTs == 0 {
		allow = true
	} else {
		allow = lastVisitorSaveTs+*seconds <= now // 距离上次访问已经过去了一段时间 才是合理的
	}
	if allow {
		err = friend.R.Set(ctx, key, now, v.visitorSaveTsExpire).Err()
	}
	return
}

func (v *visitorCtrlT) VisitorList(ctx context.Context, uid int64, req *friendpb.VisitorListReq) (list []*friend.VisitorRPC, total int64, visitorsTotal, visitorsRepeated *commonpb.CounterInt64, err error) {
	var buf []byte
	var listDao []*friend.Visitor
	var ckey = v.visitorListKey(uid)

	defer func() {
		// 用户信息每次都要重新获取（有单独的用户信息缓存）
		list, err = SetupVisitorUserInfo(ctx, listDao)
	}()

	// 仅缓存第一页（缓存全部容易有bug，且意义不大）
	if req.Page.Pn == 1 {
		buf, err = friend.R.Get(ctx, ckey).Bytes()
		if db.IgnoreNilErr(err) != nil {
			return
		}
		if buf != nil {
			var model visitorListT
			_ = jsoniter.Unmarshal(buf, &model)
			if model.Total == 0 {
				return
			}
			total = model.Total
			visitorsTotal = model.VisitorsTotal
			visitorsRepeated = model.VisitorsRepeated
			listDao = model.List
			return
		}
	}

	// 非第一页 或 缓存不存在，从数据库获取
	listDao, total, err = dao.VisitorList(ctx, uid, req.Page)
	if err != nil {
		return
	}
	if req.Page.Pn > 1 {
		return
	}

	// 第一页有特别的逻辑
	// 缓存 + 获取访客数量统计
	visitorsTotal, visitorsRepeated, err = v.getVisitorStatDao(ctx, uid)
	if err != nil {
		return
	}

	if total == 0 {
		buf = []byte("{}") // 空也缓存，避免缓存穿透
	} else {
		buf, _ = jsoniter.Marshal(&visitorListT{List: listDao, Total: total, VisitorsTotal: visitorsTotal, VisitorsRepeated: visitorsRepeated})
	}
	err = friend.R.Set(ctx, ckey, buf, v.visitorListExpire).Err()
	return
}

func (v *visitorCtrlT) getVisitorStatDao(ctx context.Context, uid int64) (ctTotal, ctRepeated *commonpb.CounterInt64, err error) {
	// 下面获取统计的【全部访问次数】
	total, err1 := dao.VisitorsTotal(ctx, uid, utime.DateToday())
	yesterdayTotal, err2 := dao.VisitorsTotal(ctx, uid, utime.DateYesterday())
	if err1 != nil || err2 != nil {
		return nil, nil, xerr.JoinErrors(err1, err2)
	}
	ctTotal = &commonpb.CounterInt64{
		Count: total,
		Delta: total - yesterdayTotal,
	}

	// 下面获取统计的【重复访问次数】
	total, err1 = dao.VisitorsRepeated(ctx, uid, utime.DateToday())
	yesterdayTotal, err2 = dao.VisitorsRepeated(ctx, uid, utime.DateYesterday())
	if err1 != nil || err2 != nil {
		return nil, nil, xerr.JoinErrors(err1, err2)
	}
	ctRepeated = &commonpb.CounterInt64{
		Count: total,
		Delta: total - yesterdayTotal,
	}
	return
}

func (v *visitorCtrlT) CleanVisitorList(ctx context.Context, uid int64) error {
	err := friend.R.Del(ctx, v.visitorListKey(uid)).Err()
	return xerr.WrapRedis(err)
}
