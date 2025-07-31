package dao

import (
	"context"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/util/ucrypto"
)

type UpdateUserInfoController struct {
}

var UpdateUserInfoCtrl = UpdateUserInfoController{}

// 更新核心表字段
type updateCoreField string

const (
	CoreAvatar      updateCoreField = "avatar"
	CoreNickname    updateCoreField = "nickname"
	CoreFirstname   updateCoreField = "firstname"
	CoreLastname    updateCoreField = "lastname"
	CoreDescription updateCoreField = "description"
	CoreBirthday    updateCoreField = "birthday"
	CoreSex         updateCoreField = "sex"
	CorePhone       updateCoreField = "phone"
	CoreEmail       updateCoreField = "email"
)

// 更新扩展表字段
type updateExtField string

const (
	ExtEducation             updateExtField = "education"
	ExtHeight                updateExtField = "height"
	ExtWeight                updateExtField = "weight"
	ExtEmotional             updateExtField = "emotional"
	ExtYearIncome            updateExtField = "year_income"
	ExtOccupation            updateExtField = "occupation"
	ExtHometown              updateExtField = "hometown"
	ExtLivingHouse           updateExtField = "living_house"
	ExtHouseBuying           updateExtField = "house_buying"
	ExtCarBuying             updateExtField = "car_buying"
	ExtUniversity            updateExtField = "university"
	ExtTags                  updateExtField = "tags"
	ExtIsRealPersonCertified updateExtField = "is_realperson_certified"
	ExtIsRealNameCertified   updateExtField = "is_realname_certified"
)

func (UpdateUserInfoController) UpdateCoreField(ctx context.Context, uid int64, field updateCoreField, content interface{}) error {
	return user.Q.WithContext(ctx).Model(user.User{}).Where("uid=?", uid).Update(string(field), content).Error
}

func (UpdateUserInfoController) UpdateExtField(ctx context.Context, uid int64, field updateExtField, content interface{}) error {
	return user.Q.WithContext(ctx).Model(user.UserExt{}).Where("uid=?", uid).Update(string(field), content).Error
}

// UpdatePassword old, new 都是未加盐的hash
func (UpdateUserInfoController) UpdatePassword(ctx context.Context, ignoreOldPass bool, uid int64, oldp, newp, salt string) (bool, error) {
	_, mod, err := GetUser(ctx, uid)
	if err != nil {
		return false, err
	}
	if mod.Id == 0 {
		return false, xerr.ErrUserNotFound
	}

	if oldp == newp {
		return false, xerr.ErrPasswdNoChange
	}
	if salt == "" {
		return false, xerr.ErrParams.New("no salt provide on update password")
	}
	// 检查密码是否变化
	if mod.Password != "" && mod.PasswdSalt != "" {
		newPasswdHashOnOldSalt, _ := ucrypto.Sha1(newp, mod.PasswdSalt)
		if newPasswdHashOnOldSalt == mod.Password {
			return false, xerr.ErrPasswdNoChange
		}
	}
	r := user.Q.WithContext(ctx).Model(user.User{}).Where("uid=?", uid)
	if !ignoreOldPass {
		oldPasswdOnOldSalt, _ := ucrypto.Sha1(oldp, mod.PasswdSalt)
		r = r.Where("password = ?", oldPasswdOnOldSalt)
	}
	newPassHash, _ := ucrypto.Sha1(newp, salt) // 新密码使用新salt
	r = r.Updates(map[string]interface{}{
		"password":      newPassHash,
		"password_salt": salt,
	})
	return r.RowsAffected == 1, r.Error
}

func (UpdateUserInfoController) ClearPassword(ctx context.Context, uid int64) error {
	return user.Q.WithContext(ctx).Model(user.User{}).Where("uid=?", uid).Updates(map[string]interface{}{
		"password":      "",
		"password_salt": "",
	}).Error
}
