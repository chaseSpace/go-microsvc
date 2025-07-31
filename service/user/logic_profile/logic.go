package logic_profile

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/bizcomm/commuser"
	"microsvc/consts"
	"microsvc/infra/svccli/rpc"
	modeluser "microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/cache"
	"microsvc/service/user/dao"
	"microsvc/util/urand"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) GetUserInfo(ctx context.Context, req *userpb.GetUserInfoReq) (*userpb.GetUserInfoRes, error) {
	var err error
	for _, id := range req.Uids {
		if id < 1 {
			err = xerr.ErrParams.New("invalid arg:`uids` contains <%d>", id)
			return nil, err
		}
	}
	caller := auth.GetAuthUID(ctx)
	var uids = req.Uids
	if req.GetCaller { // 查自己
		uids = append(uids, caller)
	}
	if len(uids) == 0 {
		return nil, xerr.ErrParams.New("No uid provide")
	}
	umap, err := cache.GetUser(ctx, uids...)
	if err != nil {
		return nil, err
	}
	rsp := &userpb.GetUserInfoRes{Umap: make(map[int64]*commonpb.User)}

	for _, uid := range req.Uids {
		if _, ok := umap[uid]; ok {
			rsp.Umap[uid] = umap[uid].ToPB()
		} else if req.PopulateNotFound {
			rsp.Umap[uid] = commuser.NewUnknownUser(uid).ToPB()
		}
	}

	if req.GetCaller {
		if _, ok := umap[caller]; ok {
			rsp.Caller = umap[caller].ToPB()
		} else { // caller 必须有信息
			rsp.Caller = commuser.NewUnknownUser(caller).ToPB()
		}
	}

	return rsp, nil
}

func (ctrl) UpdateUserInfo(ctx context.Context, req *userpb.UpdateUserInfoReq) (res *userpb.UpdateUserInfoRes, err error) {
	uid := auth.GetAuthUID(ctx)
	if len(req.BodyArray) == 0 {
		return nil, xerr.ErrParams.New("请指定更新字段")
	}
	umodel, err := cache.GetOneUser(ctx, uid)
	if err != nil {
		return
	}
	res = new(userpb.UpdateUserInfoRes)

	for _, body := range req.BodyArray {
		method, err := UpdateUserInfoCtrl.getFieldMethod(body.FieldType)
		if err != nil {
			return nil, err
		}
		if method.UpdateByUser == nil {
			return nil, xerr.ErrParams.New("不支持[用户]更新的字段: %s", body.FieldType)
		}
		if err = method.UpdateByUser(ctx, umodel, body); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (c ctrl) ResetPassword(ctx context.Context, req *userpb.ResetPasswordReq) (*userpb.ResetPasswordRes, error) {
	// 执行静态检查（后端接收的是hash，复杂性在前端校验）
	err := modeluser.InfoStaticCheckCtrl.CheckPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}
	// check email code
	check, err := rpc.Thirdparty().VerifyEmailCodeInt(ctx, &thirdpartypb.VerifyEmailCodeIntReq{
		Email:             req.Email,
		Code:              req.VerifyCode,
		Scene:             commonpb.EmailCodeScene_ECS_ResetPasswd,
		DeleteAfterVerify: true,
	})
	if err != nil {
		if xerr.ErrEmailCodeNeedSendFirst.Is(err) {
			return nil, xerr.ErrParams.New("The status of this page is no longer valid. Please re-initiate the previous procedure")
		}
		return nil, err
	}
	if !check.IsMatch {
		return nil, xerr.ErrInvalidVerifyCode.New("Wrong verification code")
	}

	_, mod, err := dao.GetUserFromTh(ctx, commonpb.SignInType_SIT_EMAIL, req.Email)
	if err != nil {
		return nil, err
	}
	if mod.Id == 0 {
		return nil, xerr.ErrUserNotFound
	}
	salt := urand.Strings(consts.PasswordSaltLen)
	if updated, err := dao.UpdateUserInfoCtrl.UpdatePassword(ctx, true, mod.Uid, "", req.NewPassword, salt); err != nil {
		return nil, err
	} else if !updated {
		//return nil, xerr.ErrParams.New("Your new password is the same as the old one")
	}
	return &userpb.ResetPasswordRes{}, nil
}
