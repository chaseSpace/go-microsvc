package implement

import (
	"context"
	"errors"
	"microsvc/bizcomm/auth"
	"microsvc/bizcomm/commuser"
	"microsvc/deploy"
	"microsvc/enums"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/base"
	"microsvc/service/user/dao"
	deploy2 "microsvc/service/user/deploy"
	"microsvc/util"
	"microsvc/util/db"
	"microsvc/util/ucrypto"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Base struct {
}

// 通用注册方法
func (b Base) commonSignUp(ctx context.Context, tx *gorm.DB, anyAccount, password, salt string, typ commonpb.SignInType, body *userpb.SignUpBody) (*user.User, error) {
	if body == nil {
		// 根据应用情况决定是否需要基本信息才能注册
		//return nil, xerr.ErrParams.New("请提供用户基本信息")
		body = &userpb.SignUpBody{Extra: &userpb.SignUpExt{}}
	}
	var err error
	var birth time.Time
	if body.Birthday != "" {
		birth, err = time.ParseInLocation(time.DateOnly, body.Birthday, time.Local)
		if err != nil {
			return nil, xerr.ErrParams.New("Invalid birthday")
		}
		if birth.After(time.Now().AddDate(-8, 0, 0)) || birth.Before(time.Now().AddDate(-90, 0, 0)) {
			return nil, xerr.ErrParams.New("Sorry, the age is not in a valid range")
		}
	}
	if body.Extra == nil {
		body.Extra = &userpb.SignUpExt{}
	}
	if body.Avatar == "" {
		body.Avatar = commuser.GetDefaultAvatar()
	}

	passwordHash, err := b.processPassword(password, salt)
	if err != nil {
		return nil, err
	}

	umodel := &user.User{
		Uid:        1,
		Nickname:   body.Nickname,
		Firstname:  body.Firstname,
		Lastname:   body.Lastname,
		Birthday:   birth,
		Sex:        enums.Sex(body.Sex),
		Avatar:     body.Avatar,
		Email:      body.Email,
		RegChannel: body.Extra.Channel,
		RegType:    typ,
		Password:   passwordHash,
		PasswdSalt: salt,
	}
	switch typ {
	case commonpb.SignInType_SIT_PHONE:
		umodel.Phone = &anyAccount
	}
	umodel.GenPartFields()
	err = umodel.Check() // 检查user各项参数
	if err != nil {
		return nil, err
	}
	umodel.Uid = 0 // reset uid

	err = Base{}.commonCreateUser(ctx, tx, umodel)
	if err != nil {
		return nil, err
	}
	return umodel, nil
}

func (Base) processPassword(passwd, salt string) (string, error) {
	if passwd != "" {
		if salt == "" {
			return "", xerr.ErrInternal.New("salt is required")
		}
		passwordHash, err := ucrypto.Sha1(passwd, salt)
		if err != nil {
			return "", xerr.ErrInternal.New("Encrypting password failed").Append(err)
		}
		return passwordHash, nil
	}
	if salt != "" {
		return "", xerr.ErrInternal.New("salt is given, but no password")
	}
	return "", nil
}

func (Base) commonCreateUser(ctx context.Context, tx *gorm.DB, userModel *user.User) (err error) {
	tryInsert := func() (duplicate bool, err error) {
		// 搜索测试函数：TestConcurrencySignUp
		_uid, err := base.UidGenerator.GenUid(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return true, nil
			}
			return false, err
		}
		uid := int64(_uid)
		userModel.Uid = uid

		xlog.Info("commonCreateUser before", zap.Any("model", userModel))

		// inserting
		err = dao.CreateUser(tx, userModel)

		tableIdx := ""
		if db.IsMysqlDuplicateErr(err, &tableIdx) {
			if tableIdx == user.TableUserUKPhone {
				return false, xerr.ErrAccountAlreadyExists
			}
			if tableIdx != user.TableUserUKUID {
				return false, xerr.ErrInternal.New("unexpected conflict idx: %s for user table", tableIdx)
			}
			return true, nil
		}
		if err != nil {
			return
		}
		err = dao.CreateUserExt(tx, &user.UserExt{Uid: uid, Tags: make([]string, 0)})
		if db.IsMysqlDuplicateErr(err, &tableIdx) {
			if tableIdx == `'PRIMARY'` { // 在 user 表刚创建的uid ，却在 user_ext 表中存在，应该是手动插入导致的，应该手动清理
				return false, xerr.ErrInternal.New("Table `user_ext` contains confusion data")
			}
			return false, xerr.ErrInternal.New("unexpected conflict idx: %s for user_ext table", tableIdx)
		}
		return false, err
	}

	var duplicate bool

	duplicate, err = tryInsert()
	if err != nil {
		return
	}

	if duplicate {
		// 注意：手动往user表插入记录可能会触发此逻辑，尝试手动删除 redis key：UidGenerator
		err = xerr.ErrInternal.New("Registration is busy. Please try again soon")
		xlog.Error("commonCreateUser 可能是注册请求并发过大，也可能是手动往user表插入了数据，导致uid重复", zap.Any("lastModel", *userModel))
		return
	}
	return
}

func (Base) GenSignToken(ctx context.Context, umodel *user.User) (string, time.Duration, error) {
	now := time.Now()

	expiry, err := deploy2.UserConf.GetTokenExpiry()
	if err != nil {
		return "", 0, err
	}
	var expiresAt *jwt.NumericDate
	if expiry > 0 {
		expiresAt = jwt.NewNumericDate(now.Add(expiry))
	}
	token, err := auth.GenerateJwT(
		&auth.SvcClaims{
			SvcCaller: auth.SvcCaller{
				Credential: auth.Credential{
					Uid:      umodel.Uid,
					Nickname: umodel.Nickname,
					Sex:      umodel.Sex,
					LoginAt:  now.Format(time.DateTime),
					RegAt:    umodel.CreatedAt.Format(time.DateTime),
				},
			},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: expiresAt,
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				Issuer:    auth.TokenIssuer,
				Subject:   cast.ToString(umodel.Uid),
				ID:        util.NewKsuid(),
			},
		}, deploy.XConf.SvcTokenSignKey)

	return token, expiry, err
}

type SignInExt struct {
	PhoneAreaCode    string
	Phone            string
	OauthCacheKey    string
	OauthUserProfile struct {
		UniqId    string // 第三方user唯一标识，用于本地注册
		AccountId string // 第三方user id，可能不唯一
		Avatar    string // 第三方user 头像地址
	}
}

// 统一登录注册接口
// 根据账号查询用户信息，查到则走登录流程，未查到则走注册流程
type __SignInImpl interface {
	CheckSignInReq(ctx context.Context, req *userpb.SignInAllReq) (*SignInExt, error) // 针对登录：特定的复杂的检查逻辑
	QueryUser(ctx context.Context, req *userpb.SignInAllReq, ext *SignInExt) (*user.User, error)
	CheckSignUpReq(ctx context.Context, req *userpb.SignUpAllReq) (*SignInExt, error) // 针对注册：特定的复杂的检查逻辑
	SignUp(ctx context.Context, req *userpb.SignUpAllReq, ext *SignInExt) (*user.User, error)
}

var SignInRegistry = map[commonpb.SignInType]__SignInImpl{
	commonpb.SignInType_SIT_PHONE:   __SignInPhone{},
	commonpb.SignInType_SIT_WX_APP:  __SignInWxApp{},
	commonpb.SignInType_SIT_WX_MINI: __SignInWxMini{},
	commonpb.SignInType_SIT_EMAIL:   __SignInEmail{},
	//commonpb.SignInType_SIT_THIRD_GOOGLE: __SignInGOOGLE{},
	commonpb.SignInType_SIT_THIRD_GITHUB: __SignInGithub{},
}
