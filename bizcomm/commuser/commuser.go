package commuser

import (
	"fmt"
	"microsvc/consts"
	"microsvc/model"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/service/user/deploy"
	"regexp"
	"strings"
)

func UserListToMap(list []*user.User) (umap map[int64]*user.User) {
	if len(list) > 0 {
		umap = make(map[int64]*user.User)
		for _, i := range list {
			umap[i.Uid] = i
		}
	}
	return
}

var supportedSmsPhoneAreaCode = map[string]struct{}{
	"1":  {}, // US
	"86": {},
}

type phoneTool struct {
}

var PhoneTool = &phoneTool{}

func (phoneTool) IsPhoneAreaCodeSupported(areaCode string) bool {
	_, ok := supportedSmsPhoneAreaCode[areaCode]
	return ok
}

func (v phoneTool) CheckPhone(areaCode, phone string) (string, error) {
	if !v.IsPhoneAreaCodeSupported(areaCode) {
		return "", xerr.ErrNotSupportedPhoneArea
	}
	if !regexp.MustCompile(`\d{5,11}`).MatchString(phone) {
		return "", xerr.ErrInvalidPhoneNo
	}
	switch true {
	case areaCode == "86":
		if len(phone) != 11 {
			return "", xerr.ErrInvalidLenPhoneNo.New("86编码请输入11位手机号 (CN locale)")
		}
	case len(areaCode) == 3: // 美国的区号是3位，但不固定
		if len(phone) < 5 {
			return "", xerr.ErrInvalidLenPhoneNo.New("The length of phone number cannot less than 5 (US locale)")
		}
	default:
		return "", xerr.ErrNotSupportedPhoneArea
	}
	return v.GetDBPhone(areaCode, phone), nil
}

func (v phoneTool) CheckPhoneStr(phoneStr string) (string, error) {
	a, p, e := v.ParsePhoneStr(phoneStr)
	if e != nil {
		return "", e
	}
	return v.CheckPhone(a, p)
}

func (phoneTool) GetDBPhone(areaCode, phone string) string {
	return fmt.Sprintf("+%s%s", areaCode, phone)
}

func (phoneTool) ParsePhoneStr(phoneStr string) (areaCode, phoneNo string, err error) {
	ss := strings.Split(phoneStr, "|")
	if len(ss) != 2 {
		return "", "", xerr.ErrInvalidPhoneNo
	}
	areaCode, phoneNo = ss[0], ss[1]
	return
}

func CheckPhoneSmsCode(code string) error {
	if len(code) != consts.PhoneSmsCodeLen {
		return xerr.ErrInvalidVerifyCode
	}
	return nil
}

func NewUnknownUser(uid int64) *user.User {
	return &user.User{
		TableBase: model.TableBase{},
		Uid:       uid,
		Nickname:  "Unknown-User",
		Firstname: "?",
		Lastname:  "?",
	}
}

func GetDefaultAvatar() string {
	return deploy.UserConf.DefaultAssets.Avatar
}
