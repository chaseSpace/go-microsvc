package user

import (
	"microsvc/enums"
	"microsvc/model"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util/urand"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cast"
)

const (
	TableUserUKUID   = "'idx_uid'"
	TableUserUKPhone = "'idx_phone'"
)

type User struct {
	model.TableBase
	Uid         int64               `gorm:"column:uid" json:"uid"` // 内部id
	Nid         *int64              `gorm:"column:nid" json:"nid"` // 靓号id
	Avatar      string              `gorm:"column:avatar" json:"avatar"`
	Nickname    string              `gorm:"column:nickname" json:"nickname"`
	Firstname   string              `gorm:"column:firstname" json:"firstname"`
	Lastname    string              `gorm:"column:lastname" json:"lastname"`
	Description string              `gorm:"column:description" json:"description"`
	Birthday    time.Time           `gorm:"column:birthday" json:"birthday"` // DB类型：date
	Sex         enums.Sex           `gorm:"column:sex" json:"sex"`
	PasswdSalt  string              `gorm:"column:password_salt" json:"password_salt"`
	Password    string              `gorm:"column:password" json:"password"` // password hash
	Phone       *string             `gorm:"column:phone" json:"phone"`
	RegChannel  string              `gorm:"column:reg_channel" json:"reg_channel"`
	RegType     commonpb.SignInType `gorm:"column:reg_type" json:"reg_type"`
	Email       string              `gorm:"column:email" json:"email"`
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) GenPartFields() {
	if u.Nickname == "" {
		u.Nickname = urand.RandName()
	}
	if u.Birthday.IsZero() {
		u.Birthday = time.Now().AddDate(2000, 1, 1)
	}
	if u.Sex.IsInvalid() {
		u.Sex = enums.SexMale
	}
}

// Check 这里仅做静态校验，不涉及数据库和rpc调用
func (u *User) Check() error {
	if !(u.Uid > 0) {
		return xerr.ErrParams.New("invalid uid")
	}
	if u.Nid != nil && *u.Nid < 0 {
		return xerr.ErrParams.New("invalid nid")
	}
	u.Nickname = strings.TrimSpace(u.Nickname)
	if err := InfoStaticCheckCtrl.CheckStringField(u.Nickname, "Nickname", 2, 20); err != nil {
		return err
	}
	//if err := InfoStaticCheckCtrl.CheckStringField(u.Firstname, "Firstname", 2, 15); err != nil {
	//	return err
	//}
	//if err := InfoStaticCheckCtrl.CheckStringField(u.Lastname, "Lastname", 2, 15); err != nil {
	//	return err
	//}
	if u.Sex.IsInvalid() {
		return xerr.ErrParams.New("Please set your gender")
	}
	if !u.Birthday.IsZero() && time.Since(u.Birthday).Hours()/(24*360) > 90 {
		return xerr.ErrParams.New("Oh, the age is too large")
	}
	//if u.Avatar == "" {
	//	return xerr.ErrParams.New("请设置头像")
	//}
	return nil
}

func (u *User) BirthdayStr() string {
	return u.Birthday.Format(time.DateOnly)
}

func (u *User) Age() int32 {
	today := time.Now()
	age := today.Year() - u.Birthday.Year()
	if today.YearDay() < u.Birthday.YearDay() {
		age--
	}
	return int32(age)
}

func (u *User) ToPB() *commonpb.User {
	return &commonpb.User{
		Uid:       u.Uid,
		Nid:       cast.ToInt64(u.Nid),
		Nickname:  u.Nickname,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Birthday:  u.Birthday.Format(time.DateOnly),
		Sex:       u.Sex.ToPB(),
		Phone:     cast.ToString(u.Phone),
		Email:     u.Email,
		Avatar:    u.Avatar,
	}
}
func (u *User) ToTinyPB() *commonpb.UserTiny {
	return &commonpb.UserTiny{
		Uid:       u.Uid,
		Nickname:  u.Nickname,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Sex:       u.Sex.ToPB(),
		Avatar:    u.Avatar,
	}
}

// InfoStaticCheckController 静态检查用户信息，公共方法，不能调用db
type InfoStaticCheckController struct {
}

var InfoStaticCheckCtrl = InfoStaticCheckController{}

func (InfoStaticCheckController) CheckStringField(v, keyName string, min, max int) error {
	l := utf8.RuneCountInString(v)
	if l < min {
		return xerr.ErrParams.New("%s is too short", keyName)
	} else if l > max {
		return xerr.ErrParams.New("%s is too long", keyName)
	}
	return nil
}

func (InfoStaticCheckController) CheckPassword(v string) error {
	// 前段传入的密码密文要求是40位字符（sha1）
	if len(v) != 40 || !regexp.MustCompile("\\w+").MatchString(v) {
		return xerr.ErrPasswdFormat
	}
	return nil
}

func (InfoStaticCheckController) CheckDescription(v string) error {
	if utf8.RuneCountInString(v) > 60 {
		return xerr.ErrParams.New("Description is too long")
	}
	return nil
}

func (InfoStaticCheckController) CheckBirthday(v string) error {
	birthday, err := time.ParseInLocation(time.DateOnly, v, time.Local)
	if err == nil {
		if !birthday.IsZero() && time.Since(birthday).Hours()/(24*360) <= 80 {
			return nil
		}
	}
	return xerr.ErrParams.New("Birthday is invalid")
}

func (InfoStaticCheckController) CheckSex(sex enums.Sex) error {
	if sex.IsInvalid() {
		return xerr.ErrParams.New("Invalid gender")
	}
	return nil
}

// UserExt 用户扩展信息结构体
type UserExt struct {
	Uid                   int64                    `gorm:"column:uid" json:"uid"`                                         // uid
	VoiceURL              string                   `gorm:"column:voice_url" json:"voice_url"`                             // 语音签名
	Education             commonpb.EducationType   `gorm:"column:education" json:"education"`                             // 学历
	Height                uint8                    `gorm:"column:height" json:"height"`                                   // 身高
	Weight                uint8                    `gorm:"column:weight" json:"weight"`                                   // 体重
	Emotional             commonpb.EmotionalType   `gorm:"column:emotional" json:"emotional"`                             // 情感状态
	YearIncome            commonpb.YearIncomeType  `gorm:"column:year_income" json:"year_income"`                         // 年收入
	Occupation            string                   `gorm:"column:occupation" json:"occupation"`                           // 职业
	Hometown              string                   `gorm:"column:hometown" json:"hometown"`                               // 籍贯/家乡
	LiveHouse             commonpb.LivingHouseType `gorm:"column:living_house" json:"living_house"`                       // 居住方式
	HouseBuying           commonpb.HouseBuyingType `gorm:"column:house_buying" json:"house_buying"`                       // 购房情况
	CarBuying             commonpb.CarBuyingType   `gorm:"column:car_buying" json:"car_buying"`                           // 购车情况
	University            string                   `gorm:"column:university" json:"university"`                           // 毕业学校
	Tags                  []string                 `gorm:"column:tags;serializer:json" json:"tags"`                       // 其他标签
	IsRealpersonCertified bool                     `gorm:"column:is_realperson_certified" json:"is_realperson_certified"` // 是否真人
	IsRealnameCertified   bool                     `gorm:"column:is_realname_certified" json:"is_realname_certified"`     // 是否实名
	// No CreatedAt, Just Use Core Table
}

func (u *UserExt) TableName() string {
	return "user_ext"
}

const (
	TableUserWxAppUKAccount = "'uk_account'"
)

type UserRegisterWeixin struct {
	model.TableBase
	Uid      int64            `gorm:"column:uid" json:"uid"`
	Account  string           `gorm:"column:account" json:"account"`
	UnionId  string           `gorm:"column:union_id" json:"union_id"`
	Nickname string           `gorm:"column:nickname" json:"nickname"`
	Type     enums.UserWxType `gorm:"column:type" json:"type"` // 含app/小程序/公众号
}

func (UserRegisterWeixin) TableName() string {
	return "user_register_weixin"
}

const (
	TableUserThUKAccountThType = "'uk_account_thtype'"
)

type UserRegisterTh struct {
	model.TableBase
	Uid     int64               `gorm:"column:uid" json:"uid"`
	Account string              `gorm:"column:account" json:"account"`
	ThType  commonpb.SignInType `gorm:"column:th_type" json:"th_type"`
}

func (UserRegisterTh) TableName() string {
	return "user_register_th"
}
