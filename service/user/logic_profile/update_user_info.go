package logic_profile

import (
	"context"
	"microsvc/bizcomm/commadmin"
	"microsvc/bizcomm/commuser"
	"microsvc/bizcomm/mq"
	"microsvc/consts"
	"microsvc/enums"
	"microsvc/infra/svccli/rpc"
	"microsvc/infra/xmq"
	modeluser "microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/cache"
	"microsvc/service/user/dao"
	"microsvc/service/user/deploy"
	"microsvc/util/ucrypto"
	"microsvc/util/ujson"
	"microsvc/util/urand"
	"strings"

	"github.com/spf13/cast"

	"github.com/pkg/errors"
)

type UpdateUserInfoController struct {
	ContextKey string
}

var UpdateUserInfoCtrl = UpdateUserInfoController{
	ContextKey: "update_type",
}

var rateConfigKeyMap = map[userpb.UserInfoType]string{
	userpb.UserInfoType_UUIT_Avatar:   commadmin.ConfigKeyUpdateRateAvatar,
	userpb.UserInfoType_UUIT_Nickname: commadmin.ConfigKeyUpdateRateNickname,
	userpb.UserInfoType_UUIT_Desc:     commadmin.ConfigKeyUpdateRateDescription,
	userpb.UserInfoType_UUIT_Password: commadmin.ConfigKeyUpdateRatePassword,
	userpb.UserInfoType_UUIT_Phone:    commadmin.ConfigKeyUpdateRatePhone,
	userpb.UserInfoType_UUIT_Sex:      commadmin.ConfigKeyUpdateRateSex,
	userpb.UserInfoType_UUIT_Birthday: commadmin.ConfigKeyUpdateRateBirthday,
}

type UpdateInfoRate struct {
	DurationLimit  string   `mapstructure:"duration_limit" json:"duration_limit"`
	DateRangeLimit []string `mapstructure:"date_range_limit" json:"date_range_limit"`
	Banned         bool     `mapstructure:"banned" json:"banned"`
	MaxHistoryLen  int64    `mapstructure:"max_history_len" json:"max_history_len"`
}

type FieldUpdateMethod struct {
	UpdateByUser  func(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error
	UpdateByAdmin func(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error
}

func (c UpdateUserInfoController) getFieldMethod(fieldTyp userpb.UserInfoType) (*FieldUpdateMethod, error) {
	method := map[userpb.UserInfoType]*FieldUpdateMethod{
		userpb.UserInfoType_UUIT_Avatar:        {UpdateByUser: c.UpdateAvatar, UpdateByAdmin: c.AdminUpdateAvatar},
		userpb.UserInfoType_UUIT_Nickname:      {UpdateByUser: c.UpdateNickname, UpdateByAdmin: c.AdminUpdateNickname},
		userpb.UserInfoType_UUIT_Firstname:     {UpdateByUser: c.UpdateFirstname, UpdateByAdmin: c.AdminUpdateFirstname},
		userpb.UserInfoType_UUIT_Lastname:      {UpdateByUser: c.UpdateLastname, UpdateByAdmin: c.AdminUpdateLastname},
		userpb.UserInfoType_UUIT_Desc:          {UpdateByUser: c.UpdateDescription, UpdateByAdmin: c.AdminUpdateDescription},
		userpb.UserInfoType_UUIT_Birthday:      {UpdateByUser: c.UpdateBirthday, UpdateByAdmin: c.AdminUpdateBirthday},
		userpb.UserInfoType_UUIT_Sex:           {UpdateByUser: c.UpdateSex, UpdateByAdmin: c.AdminUpdateSex},
		userpb.UserInfoType_UUIT_Phone:         {UpdateByUser: c.UpdatePhone, UpdateByAdmin: c.AdminUpdatePhone},
		userpb.UserInfoType_UUIT_Password:      {UpdateByUser: c.UpdatePassword, UpdateByAdmin: c.AdminUpdatePassword},
		userpb.UserInfoType_UUIT_ClearPassword: {UpdateByUser: nil, UpdateByAdmin: c.AdminClearPassword}, // 用户不可重置密码
		userpb.UserInfoType_UUIT_Email:         {UpdateByUser: c.UpdateEmail, UpdateByAdmin: nil},

		// 下面是扩展字段，没有admin方法
		userpb.UserInfoType_UUIT_Education:   {UpdateByUser: c.UpdateEducation},
		userpb.UserInfoType_UUIT_Height:      {UpdateByUser: c.UpdateHeight},
		userpb.UserInfoType_UUIT_Weight:      {UpdateByUser: c.UpdateWeight},
		userpb.UserInfoType_UUIT_Emotional:   {UpdateByUser: c.UpdateEmotional},
		userpb.UserInfoType_UUIT_YearIncome:  {UpdateByUser: c.UpdateYearIncome},
		userpb.UserInfoType_UUIT_Occupation:  {UpdateByUser: c.UpdateOccupation},
		userpb.UserInfoType_UUIT_Hometown:    {UpdateByUser: c.UpdateHometown},
		userpb.UserInfoType_UUIT_LivingHouse: {UpdateByUser: c.UpdateLivingHouse},
		userpb.UserInfoType_UUIT_HouseBuying: {UpdateByUser: c.UpdateHouseBuying},
		userpb.UserInfoType_UUIT_CarBuying:   {UpdateByUser: c.UpdateCarBuying},
		userpb.UserInfoType_UUIT_University:  {UpdateByUser: c.UpdateUniversity},
		userpb.UserInfoType_UUIT_Tags:        {UpdateByUser: c.UpdateTags},
	}[fieldTyp]
	if method == nil {
		return nil, xerr.ErrParams.New("Unsupported update field：" + fieldTyp.String())
	}
	return method, nil
}

// 执行基本的更新前检查
func (c UpdateUserInfoController) beforeUpdate(ctx context.Context, isAdminOp bool, umodel *modeluser.User, body *userpb.UpdateBody) (rate *deploy.UpdateInfoRate, err error) {
	newVal := body.AnyValue

	switch body.FieldType {
	case userpb.UserInfoType_UUIT_Avatar:
		//if umodel.Avatar == newVal && newVal != "" {
		//	return nil, xerr.ErrParams.New("Avatar not change")
		//}
		if newVal == "" { // 重置为默认头像
			body.AnyValue = commuser.GetDefaultAvatar()
		}
	case userpb.UserInfoType_UUIT_Nickname:
		//if umodel.Nickname != "" && umodel.Nickname == newVal {
		//	return nil, xerr.ErrParams.New("Nickname not change")
		//}
		// 执行静态检查
		err = modeluser.InfoStaticCheckCtrl.CheckStringField(newVal, "Nickname", 2, 20)
		if err != nil {
			return nil, err
		}
	case userpb.UserInfoType_UUIT_Firstname:
		//if umodel.Firstname != "" && umodel.Firstname == newVal {
		//	return nil, xerr.ErrParams.New("Firstname not change")
		//}
		// 执行静态检查
		err = modeluser.InfoStaticCheckCtrl.CheckStringField(newVal, "Firstname", 2, 15)
		if err != nil {
			return nil, err
		}
	case userpb.UserInfoType_UUIT_Lastname:
		//if umodel.Lastname != "" && umodel.Lastname == newVal {
		//	return nil, xerr.ErrParams.New("Lastname not change")
		//}
		// 执行静态检查
		err = modeluser.InfoStaticCheckCtrl.CheckStringField(newVal, "Lastname", 2, 15)
		if err != nil {
			return nil, err
		}
	case userpb.UserInfoType_UUIT_Desc:
		//if umodel.Description == newVal {
		//	return nil, xerr.ErrParams.New("Description not change")
		//}
		// 执行静态检查
		err = modeluser.InfoStaticCheckCtrl.CheckDescription(newVal)
		if err != nil {
			return nil, err
		}
	case userpb.UserInfoType_UUIT_Birthday:
		//if umodel.BirthdayStr() == newVal && newVal != "" {
		//	return nil, xerr.ErrParams.New("Birthday not change")
		//}
		// 执行静态检查
		err := modeluser.InfoStaticCheckCtrl.CheckBirthday(newVal)
		if err != nil {
			return nil, err
		}
	case userpb.UserInfoType_UUIT_Sex:
		newSex := enums.Sex(cast.ToInt32(newVal))
		//if umodel.Sex == newSex {
		//	return nil, xerr.ErrParams.New("Gender not change")
		//}
		// 执行静态检查
		err := modeluser.InfoStaticCheckCtrl.CheckSex(newSex)
		if err != nil {
			return nil, err
		}
	case userpb.UserInfoType_UUIT_Phone:
		_, err := commuser.PhoneTool.CheckPhoneStr(newVal)
		if err != nil {
			return nil, err
		}
	case userpb.UserInfoType_UUIT_Password:
		ss := strings.Split(newVal, "|")
		if len(ss) != 2 {
			return nil, xerr.ErrPasswdFormat
		}
		oldp, newp := ss[0], ss[1]
		// 执行静态检查（后端接收的是hash，复杂性在前端校验）
		err = modeluser.InfoStaticCheckCtrl.CheckPassword(newp)
		if err != nil {
			return nil, err
		}
		if umodel.Password != "" && !isAdminOp { // 非管理员需要验证旧密码
			oldPassHash, _ := ucrypto.Sha1(oldp, umodel.PasswdSalt)
			if oldPassHash != umodel.Password {
				return nil, xerr.ErrParams.New("The old password is wrong")
			}
			newHash, _ := ucrypto.Sha1(newp, umodel.PasswdSalt)
			if oldPassHash == newHash {
				return nil, xerr.ErrPasswdNoChange
			}
		}
	case userpb.UserInfoType_UUIT_ClearPassword:
		if umodel.Password == "" {
			return nil, xerr.ErrParams.New("The password is not set, no need to reset")
		}
	case userpb.UserInfoType_UUIT_Education:
		v := cast.ToInt32(newVal)
		if commonpb.EducationType_name[v] == "" {
			return nil, xerr.ErrParams.New("invalid edu type：" + newVal)
		}
	case userpb.UserInfoType_UUIT_Emotional:
		v := cast.ToInt32(newVal)
		if commonpb.EmotionalType_name[v] == "" {
			return nil, xerr.ErrParams.New("invalid emotional type：" + newVal)
		}
	case userpb.UserInfoType_UUIT_YearIncome:
		v := cast.ToInt32(newVal)
		if commonpb.YearIncomeType_name[v] == "" {
			return nil, xerr.ErrParams.New("invalid year-income type：" + newVal)
		}
	case userpb.UserInfoType_UUIT_LivingHouse:
		v := cast.ToInt32(newVal)
		if commonpb.LivingHouseType_name[v] == "" {
			return nil, xerr.ErrParams.New("invalid living-house type：" + newVal)
		}
	case userpb.UserInfoType_UUIT_HouseBuying:
		v := cast.ToInt32(newVal)
		if commonpb.HouseBuyingType_name[v] == "" {
			return nil, xerr.ErrParams.New("invalid house-buying type：" + newVal)
		}
	case userpb.UserInfoType_UUIT_CarBuying:
		v := cast.ToInt32(newVal)
		if commonpb.CarBuyingType_name[v] == "" {
			return nil, xerr.ErrParams.New("invalid car-buying type：" + newVal)
		}
	case userpb.UserInfoType_UUIT_Tags:
		var tags []string
		err = ujson.Unmarshal([]byte(newVal), &tags)
		if err != nil {
			return nil, xerr.ErrParams.New("invalid tag chars：" + newVal)
		}
		if len(tags) == 0 {
			body.AnyValue = `[]`
		}
	}

	rate = &deploy.UpdateInfoRate{}
	if !isAdminOp { // 用户更新 需要 获取更新频率配置
		var rateConfigKey = rateConfigKeyMap[body.FieldType]
		if rateConfigKey != "" {
			res, err := rpc.Admin().ConfigCenterGet(ctx, &adminpb.ConfigCenterGetReq{Keys: []string{rateConfigKey}})
			if err != nil {
				return nil, err
			}
			cc := res.Cmap[rateConfigKey]
			if cc != nil {
				ujson.MustUnmarshal([]byte(cc.Core.Value), rate)
			}
		}
	}
	return
}

func (c UpdateUserInfoController) afterUpdateSuccess(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	if err := cache.ClearUserInfo(ctx, umodel.Uid); err != nil {
		return err
	}
	xmq.Produce(consts.TopicUserInfoUpdate, mq.NewMsgUserInfoUpdate(&mq.UserInfoUpdateBody{
		UID:  umodel.Uid,
		Body: body,
	}))
	return nil
}

func (c UpdateUserInfoController) UpdatePassword(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}

	salt := urand.Strings(consts.PasswordSaltLen)
	ss := strings.Split(body.AnyValue, "|")
	oldPassPlain, newPassPlain := ss[0], ss[1] // beforeUpdate 中已经验证过
	// DB更新
	if updated, err := dao.UpdateUserInfoCtrl.UpdatePassword(ctx, false, umodel.Uid, oldPassPlain, newPassPlain, salt); err != nil {
		return err
	} else if !updated {
		return xerr.ErrPasswdNoChange // 一种并发的情况
	}

	// 增加更新缓存
	err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen)
	if err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) AdminUpdatePassword(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}

	salt := urand.Strings(consts.PasswordSaltLen)
	newPasswdHash := body.AnyValue
	// DB更新
	if updated, err := dao.UpdateUserInfoCtrl.UpdatePassword(ctx, true, umodel.Uid, "", newPasswdHash, salt); err != nil {
		return err
	} else if !updated {
		return xerr.ErrParams.New("密码未变化")
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) AdminClearPassword(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}
	if err = dao.UpdateUserInfoCtrl.ClearPassword(ctx, umodel.Uid); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

// UpdateAvatar 用户操作更新昵称（受到频率限制）
func (c UpdateUserInfoController) UpdateAvatar(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}

	// 需要审核？
	sw, err := readSwitchReviewUserInfo(ctx, body.FieldType)
	if err != nil {
		return err
	}
	if sw.IsOpen() {
		_, err = rpc.Thirdparty().SyncReviewImage(ctx, &thirdpartypb.SyncReviewImageReq{
			Uid:  umodel.Uid,
			Uri:  body.AnyValue,
			Type: thirdpartypb.ImageType_IT_Avatar,
		})
		return err
	}
	// 无需审核，直接更新DB
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreAvatar, body.AnyValue); err != nil {
		return err
	}

	// 增加更新缓存
	err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen)
	if err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

// AdminUpdateAvatar 管理员操作更新头像（基本不受任何限制）
func (c UpdateUserInfoController) AdminUpdateAvatar(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}

	// DB更新
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreAvatar, body.AnyValue); err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

// UpdateNickname 用户操作更新昵称（受到频率限制）
func (c UpdateUserInfoController) UpdateNickname(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}

	// 需要审核？
	sw, err := readSwitchReviewUserInfo(ctx, body.FieldType)
	if err != nil {
		return err
	}
	if sw.IsOpen() {
		_, err = rpc.Thirdparty().SyncReviewText(ctx, &thirdpartypb.SyncReviewTextReq{
			Uid:  umodel.Uid,
			Text: body.AnyValue,
			Type: thirdpartypb.TextType_TT_Nickname,
		})
		return err
	}
	// 先提交到admin后台存档，可以后审
	_, err = rpc.Admin().AddReview(ctx, &adminpb.AddReviewReq{
		Uid:  umodel.Uid,
		Type: commonpb.ReviewType_RT_Text,
		Text: body.AnyValue,
		//MediaUrls: nil,
		Status:  commonpb.ReviewStatus_RS_AIPass,
		BizType: commonpb.BizType_RBT_Nickname,
		//BizUniqId: 0,
		//ThTaskId:  "",
	})
	if err != nil {
		return err
	}
	// 先发后审
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreNickname, body.AnyValue); err != nil {
		return err
	}

	// 增加更新缓存
	err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen)
	if err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

// AdminUpdateNickname 管理员操作更新昵称（基本不受任何限制）
func (c UpdateUserInfoController) AdminUpdateNickname(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}

	// DB更新
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreNickname, body.AnyValue); err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

// UpdateFirstname 用户操作更新 Firstname（受到频率限制）
func (c UpdateUserInfoController) UpdateFirstname(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}

	// 需要审核？
	sw, err := readSwitchReviewUserInfo(ctx, body.FieldType)
	if err != nil {
		return err
	}
	if sw.IsOpen() {
		_, err = rpc.Thirdparty().SyncReviewText(ctx, &thirdpartypb.SyncReviewTextReq{
			Uid:  umodel.Uid,
			Text: body.AnyValue,
			Type: thirdpartypb.TextType_TT_Firstname,
		})
		return err
	}
	// 先提交到admin后台存档，可以后审
	_, err = rpc.Admin().AddReview(ctx, &adminpb.AddReviewReq{
		Uid:  umodel.Uid,
		Type: commonpb.ReviewType_RT_Text,
		Text: body.AnyValue,
		//MediaUrls: nil,
		Status:  commonpb.ReviewStatus_RS_AIPass,
		BizType: commonpb.BizType_RBT_Firstname,
		//BizUniqId: 0,
		//ThTaskId:  "",
	})
	if err != nil {
		return err
	}
	// 先发后审
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreFirstname, body.AnyValue); err != nil {
		return err
	}
	// 增加更新缓存
	err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen)
	if err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

// AdminUpdateFirstname 管理员操作更新 Firstname（基本不受任何限制）
func (c UpdateUserInfoController) AdminUpdateFirstname(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}

	// DB更新
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreFirstname, body.AnyValue); err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

// UpdateLastname 用户操作更新 Lastname（受到频率限制）
func (c UpdateUserInfoController) UpdateLastname(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}

	// 需要审核？
	sw, err := readSwitchReviewUserInfo(ctx, body.FieldType)
	if err != nil {
		return err
	}
	if sw.IsOpen() {
		_, err = rpc.Thirdparty().SyncReviewText(ctx, &thirdpartypb.SyncReviewTextReq{
			Uid:  umodel.Uid,
			Text: body.AnyValue,
			Type: thirdpartypb.TextType_TT_Lastname,
		})
		return err
	}
	// 先提交到admin后台存档，可以后审
	_, err = rpc.Admin().AddReview(ctx, &adminpb.AddReviewReq{
		Uid:  umodel.Uid,
		Type: commonpb.ReviewType_RT_Text,
		Text: body.AnyValue,
		//MediaUrls: nil,
		Status:  commonpb.ReviewStatus_RS_AIPass,
		BizType: commonpb.BizType_RBT_Lastname,
		//BizUniqId: 0,
		//ThTaskId:  "",
	})
	if err != nil {
		return err
	}
	// 先发后审
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreLastname, body.AnyValue); err != nil {
		return err
	}
	// 增加更新缓存
	err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen)
	if err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

// AdminUpdateLastname 管理员操作更新 Lastname（基本不受任何限制）
func (c UpdateUserInfoController) AdminUpdateLastname(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}

	// DB更新
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreLastname, body.AnyValue); err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

// UpdateDescription 用户操作更新签名（受到频率限制）
func (c UpdateUserInfoController) UpdateDescription(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}

	// 需要审核？
	sw, err := readSwitchReviewUserInfo(ctx, body.FieldType)
	if err != nil {
		return err
	}
	if sw.IsOpen() {
		_, err = rpc.Thirdparty().SyncReviewText(ctx, &thirdpartypb.SyncReviewTextReq{
			Uid:  umodel.Uid,
			Text: body.AnyValue,
			Type: thirdpartypb.TextType_TT_Desc,
		})
		return err
	}
	// 无需审核，直接更新DB
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreDescription, body.AnyValue); err != nil {
		return err
	}

	// 增加更新缓存
	err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen)
	if err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) AdminUpdateDescription(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}

	// DB更新
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreDescription, body.AnyValue); err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

// UpdateBirthday 用户操作更新生日（受到频率限制）
func (c UpdateUserInfoController) UpdateBirthday(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}

	// 生日无需审核，直接更新DB
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreBirthday, body.AnyValue); err != nil {
		return err
	}

	// 增加更新缓存
	err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen)
	if err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) AdminUpdateBirthday(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}

	// DB更新
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreBirthday, body.AnyValue); err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateSex(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}

	// 性别无需审核，直接更新DB
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreSex, body.AnyValue); err != nil {
		return err
	}

	// 增加更新缓存
	err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen)
	if err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) AdminUpdateSex(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}

	// DB更新
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreSex, body.AnyValue); err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdatePhone(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}

	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CorePhone, body.AnyValue); err != nil {
		return err
	}

	// 增加更新缓存
	err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen)
	if err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateEmail(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CoreEmail, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) AdminUpdatePhone(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	_, err := c.beforeUpdate(ctx, true, umodel, body)
	if err != nil {
		return err
	}

	// DB更新
	phone := commuser.PhoneTool.GetDBPhone(body.AnyValue, body.AnyValue)
	if err = dao.UpdateUserInfoCtrl.UpdateCoreField(ctx, umodel.Uid, dao.CorePhone, phone); err != nil {
		return err
	}

	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateEducation(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtEducation, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateHeight(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtHeight, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateWeight(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtWeight, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateEmotional(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtEmotional, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateYearIncome(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtYearIncome, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateOccupation(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtOccupation, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateHometown(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtHometown, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateLivingHouse(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtLivingHouse, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateHouseBuying(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtHouseBuying, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateCarBuying(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtCarBuying, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateUniversity(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtUniversity, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func (c UpdateUserInfoController) UpdateTags(ctx context.Context, umodel *modeluser.User, body *userpb.UpdateBody) error {
	rate, err := c.beforeUpdate(ctx, false, umodel, body)
	if err != nil {
		return err
	}

	err = cache.UserInfoUpdateCtrl.DoesAllowUpdateUserInfo(ctx, body, umodel.Uid, rate)
	if err != nil {
		return errors.Wrap(err, "DoesAllowUpdateUserInfo")
	}
	if err = dao.UpdateUserInfoCtrl.UpdateExtField(ctx, umodel.Uid, dao.ExtTags, body.AnyValue); err != nil {
		return err
	}
	if err = cache.UserInfoUpdateCtrl.AddUpdateInfoHistory(ctx, umodel.Uid, body.FieldType, rate.MaxHistoryLen); err != nil {
		return err
	}
	return c.afterUpdateSuccess(ctx, umodel, body)
}

func readSwitchReviewUserInfo(ctx context.Context, infoTyp userpb.UserInfoType) (*commadmin.SwitchItem, error) {
	// 仅列出可能需要审核的信息类型
	var smap = map[userpb.UserInfoType]commadmin.SwitchKey{
		userpb.UserInfoType_UUIT_Nickname:  commadmin.SwitchKeyReviewUserNickname,
		userpb.UserInfoType_UUIT_Firstname: commadmin.SwitchKeyReviewUserFirstname,
		userpb.UserInfoType_UUIT_Lastname:  commadmin.SwitchKeyReviewUserLastname,
		userpb.UserInfoType_UUIT_Avatar:    commadmin.SwitchKeyReviewUserIcon,
		userpb.UserInfoType_UUIT_Desc:      commadmin.SwitchKeyReviewUserDescription,
	}

	if skey := smap[infoTyp]; skey == "" {
		return nil, errors.New("Undefined user info type:" + infoTyp.String())
	} else {
		item, err := commadmin.SwitchCenterGetOne(ctx, skey)
		if err != nil {
			return nil, errors.Wrap(err, "SwitchCenterGetOne")
		}
		return item, nil
	}
}
