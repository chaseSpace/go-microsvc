package logic_profile

import (
	"context"
	"errors"
	"microsvc/bizcomm/commuser"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/cache"
	"microsvc/service/user/dao"
	"microsvc/util/db"
	"microsvc/util/urand"

	"github.com/samber/lo"
)

type intCtrl struct{}

var (
	Int = intCtrl{} // 暴露struct而不是interface，方便IDE跳转
)

func (intCtrl) GetUserInfoInt(ctx context.Context, req *userpb.GetUserInfoIntReq) (*userpb.GetUserInfoIntRes, error) {
	umap, err := cache.GetUser(ctx, lo.Uniq(req.Uids)...)
	if err != nil {
		return nil, err
	}
	rsp := &userpb.GetUserInfoIntRes{Umap: make(map[int64]*commonpb.User), Umap2: make(map[int64]*commonpb.UserTiny)}
	if req.GetTiny {
		for _, id := range req.Uids {
			if v := umap[id]; v != nil {
				rsp.Umap2[id] = v.ToTinyPB()
			} else if req.PopulateNotfound {
				rsp.Umap2[id] = commuser.NewUnknownUser(id).ToTinyPB()
			}
		}
		return rsp, nil
	}
	for _, id := range req.Uids {
		if v := umap[id]; v != nil {
			rsp.Umap[id] = v.ToPB()
		} else if req.PopulateNotfound {
			rsp.Umap[id] = commuser.NewUnknownUser(id).ToPB()
		}
	}
	return rsp, nil
}

func (u intCtrl) AllocateUserNid(ctx context.Context, req *userpb.AllocateUserNidReq) (*userpb.AllocateUserNidRes, error) {
	umap, err := cache.GetUser(ctx, req.Uid)
	if err != nil {
		return nil, err
	}
	if len(umap) == 0 {
		return nil, xerr.ErrUserNotFound
	}
	var setNid *int64
	if req.Nid > 0 {
		setNid = &req.Nid
	}
	err = dao.UpdateUserNid(ctx, req.Uid, setNid)
	if db.IsMysqlDuplicateErr(err, nil) {
		err = errors.New("靓号已被使用")
	}
	return new(userpb.AllocateUserNidRes), err
}

func (u intCtrl) AdminUpdateUserInfo(ctx context.Context, req *userpb.AdminUpdateUserInfoReq) (res *userpb.AdminUpdateUserInfoRes, err error) {
	if len(req.BodyArray) == 0 {
		return nil, xerr.ErrParams.New("请指定更新字段")
	}
	umodel, err := cache.GetOneUser(ctx, req.Uid)
	if err != nil {
		return
	}
	res = new(userpb.AdminUpdateUserInfoRes)

	for _, body := range req.BodyArray {
		method, err := UpdateUserInfoCtrl.getFieldMethod(body.FieldType)
		if err != nil {
			return nil, err
		}
		if err = method.UpdateByAdmin(ctx, umodel, body); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (c intCtrl) ReviewProfile(ctx context.Context, req *userpb.ReviewProfileReq) (res *userpb.ReviewProfileRes, err error) {
	res = new(userpb.ReviewProfileRes)
	if req.IsPass { // 用户资料都是先发后审，所以通过的情况就忽略
		return
	}
	umodel, err := cache.GetOneUser(ctx, req.Uid)
	if err != nil {
		return
	}

	// 下面都是处理违规的情况
	switch req.BizType {
	case commonpb.BizType_RBT_Nickname:
		err = UpdateUserInfoCtrl.AdminUpdateNickname(ctx, umodel, &userpb.UpdateBody{
			FieldType: userpb.UserInfoType_UUIT_Nickname,
			AnyValue:  "违规昵称_" + urand.Strings(4, true),
		})
		// TODO 发送违规通知
	case commonpb.BizType_RBT_UserDesc:
		err = UpdateUserInfoCtrl.AdminUpdateDescription(ctx, umodel, &userpb.UpdateBody{
			FieldType: userpb.UserInfoType_UUIT_Desc,
			AnyValue:  "客服提醒：违规简介，请重新填写",
		})
	case commonpb.BizType_RBT_Avatar:
		err = UpdateUserInfoCtrl.AdminUpdateAvatar(ctx, umodel, &userpb.UpdateBody{
			FieldType: userpb.UserInfoType_UUIT_Avatar,
			AnyValue:  commuser.GetDefaultAvatar(),
		})
	case commonpb.BizType_RBT_Album:
	// TODO 未支持相册；给违规图打标记
	default:
		return nil, xerr.ErrParams.New("未知BizType: " + req.BizType.String())
	}
	return
}
